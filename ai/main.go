// Copyright (c) 2020, The SpeedyChess Contributors. All rights reserved.

package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"math/rand"
	"time"

	"github.com/TheDiscordian/speedychess/chess"
	"github.com/TheDiscordian/speedychess/chesspb"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
)

const (
	LOOKAHEAD        = 2
	ADDR             = "localhost:8181"
	DEBUG            = true
	WRITER_MAXBUFFER = 180               //how many packets to queue before dropping the connection
	READER_MAXWAIT   = 120 * time.Second //max time to receive no full packet from client
)

var (
	C       *chesspb.Client
	Game    *chess.Chessboard
	Playern int
	Black   bool
	MyTurn  bool
	LastMove time.Time
	DoingGuess bool
)

// FIXME AI will always promote queen

func needsPromotion(x, y int8, board *chess.Chessboard) bool {
	if (y == 7 || y == 0) && (board.Board[y][x] == chess.WhitePawn || board.Board[y][x] == chess.BlackPawn) {	
		return true
	}
	return false
}

// doGuess makes a guess at the next best move from x, y, then returns the change in score.
func doGuess(x, y int8, board *chess.Chessboard) (int, error) {
	if board.Board[y][x] == 0 {
		log.Panicln("doGuess on empty space")
	}
	Moves, EnPassantKill, CanCastleLeft, CanCastleRight := board.PossibleMoves(x, y)

	var (
		BestScore int
		BestMoves, BestEnPassantKills [][2]int8
		BestCastleLeft bool
		BestCastleRight bool
	)
	BestScore = -500
	
	BestMoves = make([][2]int8, 0, 1)
	BestEnPassantKills = make([][2]int8, 0, 1)
	
	black := chess.IsBlack(board.Board[y][x])
	
	theirScore := board.TotalValue(!black)
	
	for _, move := range Moves {
		var totalScore int
		tboard := new(chess.Chessboard)
		*tboard = *board
		val := chess.Value(tboard.Board[y][x])
		tboard.DoMove([2]int8{x, y}, move) 
		if needsPromotion(move[0], move[1], tboard) {
			totalScore += 9
			if black {
				tboard.PromotePawn(move[0], move[1], chess.BlackQueen)
			} else {
				tboard.PromotePawn(move[0], move[1], chess.WhiteQueen)
			}
		}
		if tboard.IsCheckmated(!black) {
			totalScore += 50
		} else {
			totalScore += theirScore-tboard.TotalValue(!black)
			if tboard.Threat(!black)[move[1]][move[0]] {
				totalScore -= val
			}
		}
		if totalScore == BestScore {
			BestMoves = append(BestMoves, move)
		} else if totalScore > BestScore {
			BestMoves = [][2]int8{move}
			BestScore = totalScore
		}
	}
	for _, move := range EnPassantKill {
		var totalScore int
		tboard := *board
		val := chess.Value(tboard.Board[y][x])
		tboard.DoEnPassant([2]int8{x, y}, move)
		if tboard.IsCheckmated(!black) {
			totalScore = 50
		} else {
			totalScore = theirScore-tboard.TotalValue(!black)
			if tboard.Threat(!black)[move[1]][move[0]] {
				totalScore -= val
			}
		}
		if totalScore == BestScore {
			BestEnPassantKills = append(BestEnPassantKills, move)
		} else if totalScore > BestScore {
			BestMoves = nil
			BestEnPassantKills = [][2]int8{move}
			BestScore = totalScore
		}
	}
	if CanCastleLeft {
		tboard := *board
		tboard.DoCastle([2]int8{x, y}, true)
		totalScore := theirScore-tboard.TotalValue(!black)
		if totalScore == BestScore {
			BestCastleLeft = true
		} else if totalScore > BestScore {
			BestMoves = nil
			BestEnPassantKills = nil
			BestCastleLeft = true
			BestScore = totalScore
		}
	}
	if CanCastleRight {
		tboard := *board
		tboard.DoCastle([2]int8{x, y}, false)
		totalScore := theirScore-tboard.TotalValue(!black)
		if totalScore == BestScore {
			BestCastleRight = true
		} else if totalScore > BestScore {
			BestMoves = nil
			BestEnPassantKills = nil
			BestCastleLeft = false
			BestCastleRight = true
			BestScore = totalScore
		}
	}
	
	if len(BestMoves) == 0 && len(BestEnPassantKills) == 0 && !BestCastleLeft && !BestCastleRight {
		if Game != nil && board == Game {
			log.Println("board is Game, Game is not nil, but no best moves...")
		}
		return 0, errors.New("No moves.")
	}
	
	for {
		switch n := rand.Intn(len(BestMoves)+3); {
			case n < len(BestMoves):
				if board == Game {
					log.Println("Move from: ", [2]int8{x, y}, ", to: ", BestMoves[n])
					C.Send(&chesspb.Move{Fx: uint32(x), Fy: uint32(y), Tx: uint32(BestMoves[n][0]), Ty: uint32(BestMoves[n][1]), MoveType: chesspb.Move_MoveType(0)})
				} else {
					board.DoMove([2]int8{x, y}, BestMoves[n])
				}
				return BestScore, nil
			case n == len(BestMoves) && len(BestEnPassantKills) > 0:
				if board == Game {
					log.Println("En passant move from: ", [2]int8{x, y}, ", to: ", BestEnPassantKills[0])
					C.Send(&chesspb.Move{Fx: uint32(x), Fy: uint32(y), Tx: uint32(BestEnPassantKills[0][0]), Ty: uint32(BestEnPassantKills[0][1]), MoveType: chesspb.Move_MoveType(1)})
				} else {
					board.DoEnPassant([2]int8{x, y}, BestEnPassantKills[0])
				}
				return BestScore, nil
			case n == len(BestMoves)+1 && BestCastleLeft:
				if board == Game {
					log.Println("Move from: ", [2]int8{x, y}, "castle left")
					C.Send(&chesspb.Move{Fx: uint32(x), Fy: uint32(y), MoveType: chesspb.Move_MoveType(2)})
				} else {
					board.DoCastle([2]int8{x, y}, true)
				}
				return BestScore, nil
			case n == len(BestMoves)+2 && BestCastleRight:
				if board == Game {
					log.Println("Move from: ", [2]int8{x, y}, "castle right")
					C.Send(&chesspb.Move{Fx: uint32(x), Fy: uint32(y), MoveType: chesspb.Move_MoveType(2)})
				} else {
					board.DoCastle([2]int8{x, y}, false)
				}
				return BestScore, nil
		}
	}
}

// guessBestMove iterates over every possible move ahead times, and returns a single "best move" based on score.
func guessBestMove(board *chess.Chessboard, ahead int, black bool) ([2]int8, int, error) {
	var (
		BestScore int
		BestMoves [][2]int8
	)
	BestScore = -5000
	
	for y, row := range board.Board {
		for x, p := range row {
			if p == 0 || chess.IsBlack(p) != black {
				continue
			}
			var (
				Score, EnemyScore int
				err error
				move [2]int8
			)
			tboard := new(chess.Chessboard)
			*tboard = *board
			_, err = doGuess(int8(x), int8(y), tboard)
			if err != nil { // we can't move this piece
				continue
			}
			Score = tboard.TotalValue(black) - Game.TotalValue(black) - 10
			EnemyScore = tboard.TotalValue(!black) - Game.TotalValue(!black)
			if (Score != -10 || EnemyScore != 0) && ahead == LOOKAHEAD {
				log.Println("Best score:", BestScore, "Score:", Score, "Enemy Score:", EnemyScore)
				log.Println("Move from:", x, y)
			}
			move = [2]int8{int8(x), int8(y)}
			if ahead > 0 {
				enemymove, enemyBestScore, err := guessBestMove(tboard, ahead-1, !black)
				if err == nil {
					_, err = doGuess(enemymove[0], enemymove[1], tboard)
					if err == nil {
						Score = (tboard.TotalValue(black) - board.TotalValue(black))*ahead
						Score -= enemyBestScore
					} else {
						log.Println("doGuess has no moves.")
						Score *= ahead
					}
				} else {
					log.Println("guessBestMove has no moves.")
					Score *= ahead
				}
				Score = Score - (tboard.TotalValue(!black) - board.TotalValue(!black))*ahead
			} else {
				Score -= EnemyScore
			}
			if tboard.IsCheckmated(!black) {
				/*if black == Black {
					log.Println("Possible win detected...")
				}*/
				Score = 50*(ahead+1)
			} else if tboard.IsCheckmated(black) {
				/*if black == Black {
					log.Println("Possible loss detected...")
				}*/
				Score = -50*(ahead+1)
			}
			
			if Score == BestScore {
				BestMoves = append(BestMoves, move)
			} else if Score > BestScore {
				BestMoves = [][2]int8{move}
				BestScore = Score
			}
			if ahead == LOOKAHEAD && Score != -10 {
				log.Println("Best score:", BestScore, "Score:", Score)
				log.Println("Move from:", move)
			}
		}
	}
	
	if len(BestMoves) > 0 {
		return BestMoves[rand.Intn(len(BestMoves))], BestScore, nil
	}
	return [2]int8{0, 0}, 0, errors.New("No best move found.")
}

func connect() error {
	ctx := context.Background()
	log.Println("Connecting to", ADDR, "...")
	c, _, err := websocket.Dial(ctx, "ws://"+ADDR, nil)

	if err != nil {
		log.Println("Failed to connect to server: ", err)
		return err
	}
	defer func() {
		c.Close(websocket.StatusInternalError, "the sky is falling")
	}()
	conn := websocket.NetConn(ctx, c, websocket.MessageBinary)

	C = new(chesspb.Client)
	C.W = make(chan []byte, WRITER_MAXBUFFER)
	go C.Writer(conn)

	reader := bufio.NewReader(conn) //reader for the connection
	var msg proto.Message

	log.Println("Connected!")

	for {
		last := time.Now()

		if Playern == 0 {
			C.Send(&chesspb.Join{Player: true})
		}
		if Game != nil && MyTurn && time.Since(LastMove) >= time.Millisecond * 75 && !DoingGuess {
				DoingGuess = true
				go func() {
					log.Println("Guess begin...")
					guess, _, err := guessBestMove(Game, LOOKAHEAD, Black)
					if err != nil {
						log.Println(err)
					} else {
						log.Println("Piece to move:", guess)
						doGuess(guess[0], guess[1], Game)
						log.Println("Guessed.")
					}
					LastMove = time.Now()
				}()
		}

		conn.SetReadDeadline(time.Now().Add(READER_MAXWAIT))
		err := chesspb.ReadMessage(reader, &msg) //read message into msg
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("Disconnected from server (inactive time: %.2fs)\n", time.Since(last).Seconds())
				log.Println(conn.RemoteAddr().String(), err)
			}
			return err
		}

		switch v := msg.(type) {
		case *chesspb.Promote:
			if v.To == 0 {
				if Black {
					C.Send(&chesspb.Promote{X: v.X, Y: v.Y, To: int32(chess.BlackQueen)})
				} else {
					C.Send(&chesspb.Promote{X: v.X, Y: v.Y, To: int32(chess.WhiteQueen)})
				}
			} else {
				Game.PromotePawn(int8(v.X), int8(v.Y), chess.Piece(v.To))
			}
		case *chesspb.Ping:
			//LogToConsole("[DEBUG] Ping!")
			C.Send(v)
		case *chesspb.OpponentJoined:
			log.Println("Opponent joined, game is ready to begin!")
			C.Send(new(chesspb.NewGame))
		case *chesspb.Player:
			if v.One {
				if DEBUG {
					log.Println("Joined as player 1.")
				}
				Playern = 1
			} else {
				if DEBUG {
					log.Println("Joined as player 2.")
				}
				Playern = 2
			}
		case *chesspb.Team:
			if v.Black {
				Black = true
				log.Println("Assigned to black.")
				MyTurn = false
			} else {
				Black = false
				log.Println("Assigned to white.")
				MyTurn = true
			}
			Game = chess.NewChessboard()
		case *chesspb.Move:
			switch chess.MoveType(v.MoveType) {
			case chess.RegularMove:
				Game.DoMove([2]int8{int8(v.Fx), int8(v.Fy)}, [2]int8{int8(v.Tx), int8(v.Ty)})
			case chess.EnPassant:
				Game.DoEnPassant([2]int8{int8(v.Fx), int8(v.Fy)}, [2]int8{int8(v.Tx), int8(v.Ty)})
			case chess.CastleLeft:
				Game.DoCastle([2]int8{int8(v.Fx), int8(v.Fy)}, true)
			case chess.CastleRight:
				Game.DoCastle([2]int8{int8(v.Fx), int8(v.Fy)}, false)
			}
			DoingGuess = false
			MyTurn = !MyTurn
			time.Sleep(50*time.Millisecond)
		case *chesspb.GameComplete:
			switch v.Result {
			case chesspb.GameComplete_Stalemate:
				log.Println("Game complete! Stalemate.")
			case chesspb.GameComplete_WhiteWin:
				log.Println("Game complete! White wins!")
			case chesspb.GameComplete_BlackWin:
				log.Println("Game complete! Black wins!")
			}
			// TODO STORE DATA IN DB.
			Game = nil
			Playern = 0
		case *chesspb.OpponentLeft:
			log.Println("Opponent left, need to rejoin.")
			Game = nil
			Playern = 0
		case *chesspb.Error:
			log.Println("Server error: " + v.Msg)
		}
	}
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.Println("Booting AI with look ahead of", LOOKAHEAD)
	connect()
}
