// Copyright (c) 2020, The SpeedyChess Contributors. All rights reserved.

package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/TheDiscordian/speedychess/chess"
	"github.com/TheDiscordian/speedychess/chesspb"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
)

const (
	WRITER_MAXBUFFER = 180               //how many packets to queue before dropping the connection
	READER_MAXWAIT   = 120 * time.Second //max time to receive no full packet from client
	REQS_TIME        = 1 * time.Second   //
	REQS_MAX         = 9                 //max requests allowed within REQS_TIME
	PING_INTERVAL    = 10 * time.Second
)

type Color int

const (
	None Color = iota - 1
	White
	Black
)

var (
	players     *int32 // players represents how many actual players are joined (0, 1, or 2)
	GameRunning bool
	Game        *chess.Chessboard

	BlackClient   *chesspb.Client
	WhiteClient   *chesspb.Client
	BlackMove     bool  // if true, it's black's move
	NeedPromotion Color // represents a colour that needs to promote a pawn for the game to continue
)

// TODO keep observers alive
func keepAlive() {
	for {
		if WhiteClient != nil {
			WhiteClient.Send(new(chesspb.Ping))
		}
		if BlackClient != nil {
			BlackClient.Send(new(chesspb.Ping))
		}
		time.Sleep(PING_INTERVAL)
	}
}

func handleConnection(conn net.Conn) {
	var (
		msg   proto.Message
		color Color
	)
	reader := bufio.NewReader(conn) //reader for the connection

	c := &chesspb.Client{W: make(chan []byte, 5)}
	go c.Writer(conn)

	playern := -1 // 0 or 1 are players, -1 is not assigned, anything above 1 is a spectator (and it doesn't matter)

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Client", conn.RemoteAddr().String(), "crashed goroutine:", r, "\n"+string(debug.Stack()))
		}

		if playern == 0 || playern == 1 {
			atomic.StoreInt32(players, 0)
			GameRunning = false
			BlackMove = false
			if color == Black && WhiteClient != nil {
				WhiteClient.Send(new(chesspb.OpponentLeft))
			} else if color == White && BlackClient != nil {
				BlackClient.Send(new(chesspb.OpponentLeft))
			}
			WhiteClient = nil
			BlackClient = nil
		}

		//cleanup
		conn.Close()
		c.SendBytes(nil)
	}()

	reqs := 0
	reqs_last := time.Now()

	for {
		if REQS_TIME <= time.Since(reqs_last) {
			reqs_last = time.Now()
			reqs = 0
		}
		if reqs > REQS_MAX {
			fmt.Println("Too many requests from", conn.RemoteAddr().String())
			break
		}

		conn.SetReadDeadline(time.Now().Add(READER_MAXWAIT))
		err := chesspb.ReadMessage(reader, &msg) //read message into msg
		if err != nil {
			if err.Error() != "EOF" {
				fmt.Println(conn.RemoteAddr().String(), err)
			}
			break
		}
		if atomic.LoadInt32(players) == 0 {
			playern = -1
		}

		reqs++

		switch v := msg.(type) {
		case *chesspb.Ping:
			continue
		case *chesspb.Join:
			if v.Player {
				if playern >= 0 {
					c.Send(&chesspb.Error{Msg: "You're already a player."})
					continue
				}
				if atomic.CompareAndSwapInt32(players, 0, 1) {
					playern = 0
					if rand.Intn(2) == 0 { // Randomly assign who is white or black
						WhiteClient = c
						color = White
					} else {
						BlackClient = c
						color = Black
					}
					c.Send(&chesspb.Player{One: true})
				} else if atomic.CompareAndSwapInt32(players, 1, 2) {
					playern = 1
					if WhiteClient == nil {
						WhiteClient = c
						color = White
					} else {
						BlackClient = c
						color = Black
					}
					c.Send(&chesspb.Player{One: false})
				} else {
					c.Send(&chesspb.Error{Msg: "All player slots filled."})
				}
			} else {
				playern = 3 // spectator
			}
		case *chesspb.NewGame:
			if playern == 0 {
				if BlackClient == nil || WhiteClient == nil {
					c.Send(&chesspb.Error{Msg: "Need 2 players to play."})
					continue
				}
				if !GameRunning {
					// TODO announce new game to spectators
					GameRunning = true
					Game = chess.NewChessboard()
					BlackClient.Send(&chesspb.Team{Black: true})
					WhiteClient.Send(&chesspb.Team{Black: false})
				} else {
					c.Send(&chesspb.Error{Msg: "Game already started."})
				}
			} else {
				c.Send(&chesspb.Error{Msg: "Only player 1 can start a game."})
			}
		case *chesspb.Promote:
			if !GameRunning {
				c.Send(&chesspb.Error{Msg: "Game has not started."})
				continue
			}
			if NeedPromotion != color {
				c.Send(&chesspb.Error{Msg: "You're not ready for a promotion yet."})
				continue
			}
			if !Game.PromotePawn(int8(v.X), int8(v.Y), chess.Piece(v.To)) {
				c.Send(&chesspb.Error{Msg: "Invalid selection."})
				continue
			}
			NeedPromotion = None
			WhiteClient.Send(v)
			BlackClient.Send(v)
		case *chesspb.Move:
			if !GameRunning {
				c.Send(&chesspb.Error{Msg: "Game has not started."})
				continue
			}
			if NeedPromotion != None {
				if NeedPromotion == Black {
					c.Send(&chesspb.Error{Msg: "Waiting on black to pick a promotion."})
					continue
				} else {
					c.Send(&chesspb.Error{Msg: "Waiting on white to pick a promotion."})
					continue
				}
			}
			if (color != Black && BlackMove) || (color != White && !BlackMove) {
				c.Send(&chesspb.Error{Msg: "It's not your turn."})
				continue
			}
			if Game.Board[int(v.Fy)][int(v.Fx)] == 0 {
				c.Send(&chesspb.Error{Msg: "There's no piece there."})
				continue
			}
			if (!chess.IsBlack(Game.Board[int(v.Fy)][int(v.Fx)]) && color == Black) || (chess.IsBlack(Game.Board[int(v.Fy)][int(v.Fx)]) && color == White) {
				c.Send(&chesspb.Error{Msg: "That's not your piece."})
				continue
			}
			if !Game.IsLegal([2]int8{int8(v.Fx), int8(v.Fy)}, [2]int8{int8(v.Tx), int8(v.Ty)}, chess.MoveType(v.MoveType)) {
				c.Send(&chesspb.Error{Msg: "That's not legal."})
				continue
			}
			switch chess.MoveType(v.MoveType) {
			case chess.RegularMove:
				Game.DoMove([2]int8{int8(v.Fx), int8(v.Fy)}, [2]int8{int8(v.Tx), int8(v.Ty)})
				if (v.Ty == 7 || v.Ty == 0) && (Game.Board[v.Ty][v.Tx] == chess.WhitePawn || Game.Board[v.Ty][v.Tx] == chess.BlackPawn) {
					black := chess.IsBlack(Game.Board[v.Ty][v.Tx])
					if black {
						NeedPromotion = Black
					} else {
						NeedPromotion = White
					}
					if black {
						BlackClient.Send(new(chesspb.Promote))
					} else {
						WhiteClient.Send(new(chesspb.Promote))
					}
				}
			case chess.EnPassant:
				Game.DoEnPassant([2]int8{int8(v.Fx), int8(v.Fy)}, [2]int8{int8(v.Tx), int8(v.Ty)})
			case chess.CastleLeft:
				Game.DoCastle([2]int8{int8(v.Fx), int8(v.Fy)}, true)
			case chess.CastleRight:
				Game.DoCastle([2]int8{int8(v.Fx), int8(v.Fy)}, false)
			}
			WhiteClient.Send(v)
			BlackClient.Send(v)
			BlackMove = !BlackMove
		}
	}
}

func init() {
	players = new(int32)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			fmt.Println(err)
			return
		}
		defer c.Close(websocket.StatusInternalError, "the sky is falling")

		handleConnection(websocket.NetConn(context.Background(), c, websocket.MessageBinary))

		c.Close(websocket.StatusNormalClosure, "")
	})

	go keepAlive()
	err := http.ListenAndServe(":8181", fn)
	if err != nil {
		panic(err)
	}
}
