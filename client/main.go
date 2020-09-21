// Copyright (c) 2020, The SpeedyChess Contributors. All rights reserved.

package main

import (
	"bufio"
	"context"
	"fmt"
	"syscall/js"
	"time"
	"unicode/utf8"

	"github.com/TheDiscordian/speedychess/chess"
	"github.com/TheDiscordian/speedychess/chesspb"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
)

var (
	C          *chesspb.Client
	Game       *chess.Chessboard
	Black      bool
	MyTurn     bool
	Promotion  *[2]int8
	StoredMove []int8 // nil or len(2)
)

const (
	WRITER_MAXBUFFER = 180               //how many packets to queue before dropping the connection
	READER_MAXWAIT   = 120 * time.Second //max time to receive no full packet from client
)

func LogToConsole(txt string) {
	window := js.Global()
	document := window.Get("document")

	console := document.Call("getElementById", "console")
	console.Set("innerText", txt+"\n"+console.Get("innerText").String())
}

func drawBoard(flip bool) string {
	output := make([]rune, 0, 64*12+10)
	output = append(output, []rune("<br>")...)
	brown := false

	createTile := func(x, y int, r chess.Piece) {
		if !brown {
			output = append(output, []rune(fmt.Sprintf(`<td id="%dx%d" class="notcoloured" onclick="selectpiece(%d,%d)">`, x, y, x, y))...)
		} else {
			output = append(output, []rune(fmt.Sprintf(`<td id="%dx%d" class="coloured" onclick="selectpiece(%d,%d)">`, x, y, x, y))...)
		}
		if r > 0 {
			output = append(output, []rune(fmt.Sprintf(`<span id="%s"></span>`, string(r)))...)
		}
		output = append(output, []rune("</td>")...)
		brown = !brown
	}

	if !flip {
		for y, line := range Game.Board {
			output = append(output, []rune("<tr>")...)
			for x, r := range line {
				createTile(x, y, r)
			}
			brown = !brown
			output = append(output, []rune("</tr>")...)
		}
	} else {
		for y := len(Game.Board) - 1; y >= 0; y-- {
			output = append(output, []rune("<tr>")...)
			for x := len(Game.Board) - 1; x >= 0; x-- {
				createTile(x, y, Game.Board[y][x])
			}
			brown = !brown
			output = append(output, []rune("</tr>")...)
		}
	}
	return "<table style=\"margin:-1em auto;padding-top:0px;cursor:pointer;table-layout: fixed;\">" + string(output) + "<br></table>"
}

func selectPromotion(this js.Value, args []js.Value) interface{} {
	if Promotion != nil {
		r, _ := utf8.DecodeRuneInString(args[0].String())
		C.Send(&chesspb.Promote{X: uint32(Promotion[0]), Y: uint32(Promotion[1]), To: r})
	}
	return nil
}

func selectPiece(this js.Value, args []js.Value) interface{} {
	window := js.Global()
	document := window.Get("document")

	var movet chess.MoveType

	if Game == nil || !MyTurn {
		return nil
	}
	x, y := int8(args[0].Int()), int8(args[1].Int())
	tile := document.Call("getElementById", fmt.Sprintf("%dx%d", x, y))
	red := tile.Get("red").Truthy()
	redEnPassant := tile.Get("red-enpassant").Truthy()
	redCastleLeft := tile.Get("red-castleleft").Truthy()
	redCastleRight := tile.Get("red-castleright").Truthy()
	if (red || redEnPassant || redCastleLeft || redCastleRight) && StoredMove != nil {
		var modifier int8
		if redEnPassant {
			movet = chess.EnPassant
			if Black {
				modifier = -1
			} else {
				modifier = 1
			}
		} else if redCastleLeft {
			movet = chess.CastleLeft
		} else if redCastleRight {
			movet = chess.CastleRight
		}
		C.Send(&chesspb.Move{Fx: uint32(StoredMove[0]), Fy: uint32(StoredMove[1]), Tx: uint32(x), Ty: uint32(y + modifier), MoveType: chesspb.Move_MoveType(movet)})
		StoredMove = nil
		document.Call("getElementById", "chessboard").Set("innerHTML", drawBoard(Black))
		return nil
	}
	piece := Game.Board[y][x]
	document.Call("getElementById", "chessboard").Set("innerHTML", drawBoard(Black))
	if piece == 0 || chess.IsBlack(piece) != Black {
		return nil
	}
	StoredMove = []int8{x, y}

	moves, enpassant, castleleft, castleright := Game.PossibleMoves(x, y)
	for _, move := range moves {
		square := document.Call("getElementById", fmt.Sprintf("%dx%d", move[0], move[1]))
		square.Set("style", "background-color:red;border:1px dashed;")
		square.Set("red", true)
	}
	for _, move := range enpassant {
		var modifier int8
		if Black {
			modifier = 1
		} else {
			modifier = -1
		}
		square := document.Call("getElementById", fmt.Sprintf("%dx%d", move[0], move[1]+modifier))
		square.Set("style", "background-color:red;border:1px dashed;")
		square.Set("red-enpassant", true)
	}
	if castleleft {
		square := document.Call("getElementById", fmt.Sprintf("%dx%d", x-2, y))
		square.Set("style", "background-color:red;border:1px dashed;")
		square.Set("red-castleleft", true)
	}
	if castleright {
		square := document.Call("getElementById", fmt.Sprintf("%dx%d", x+2, y))
		square.Set("style", "background-color:red;border:1px dashed;")
		square.Set("red-castleright", true)
	}
	return nil
}

func newGame(this js.Value, args []js.Value) interface{} {
	C.Send(new(chesspb.NewGame))
	return nil
}

func joinGame(this js.Value, args []js.Value) interface{} {
	C.Send(&chesspb.Join{Player: true})
	return nil
}

func connect(this js.Value, args []js.Value) interface{} {
	go func() { // blocks, so needs to be in a goroutine
		Game = nil
		MyTurn = false

		window := js.Global()
		document := window.Get("document")
		document.Call("getElementById", "connect").Set("disabled", true)

		LogToConsole("Connecting...")

		ctx := context.Background()
		c, _, err := websocket.Dial(ctx, "ws://"+document.Call("getElementById", "serveraddr").Get("value").String(), nil)
		if err != nil {
			LogToConsole(fmt.Sprintf("Failed to connect to server: %v.", err))
			document.Call("getElementById", "connect").Set("disabled", false)
			return
		}
		defer func() {
			c.Close(websocket.StatusInternalError, "the sky is falling")
			document.Call("getElementById", "connect").Set("disabled", false)
			document.Call("getElementById", "join").Set("disabled", true)
			document.Call("getElementById", "newgame").Set("disabled", true)
		}()
		conn := websocket.NetConn(ctx, c, websocket.MessageBinary)

		C = new(chesspb.Client)
		C.W = make(chan []byte, WRITER_MAXBUFFER)
		go C.Writer(conn)

		reader := bufio.NewReader(conn) //reader for the connection
		var msg proto.Message

		LogToConsole("Connected!")
		document.Call("getElementById", "join").Set("disabled", false)

		for {
			last := time.Now()
			conn.SetReadDeadline(time.Now().Add(READER_MAXWAIT))
			err := chesspb.ReadMessage(reader, &msg) //read message into msg
			if err != nil {
				if err.Error() != "EOF" {
					LogToConsole(fmt.Sprintf("Disconnected from server (inactive time: %.2fs)", time.Since(last).Seconds()))
					fmt.Println(conn.RemoteAddr().String(), err)
				}
				break
			}

			switch v := msg.(type) {
			case *chesspb.Promote:
				if v.To == 0 {
					Promotion = &[2]int8{int8(v.X), int8(v.Y)}
					if Black {
						document.Call("getElementById", "blackpromotion").Set("hidden", false)
					} else {
						document.Call("getElementById", "whitepromotion").Set("hidden", false)
					}
				} else {
					Game.PromotePawn(int8(v.X), int8(v.Y), chess.Piece(v.To))
					Promotion = nil
					document.Call("getElementById", "blackpromotion").Set("hidden", true)
					document.Call("getElementById", "whitepromotion").Set("hidden", true)
					document.Call("getElementById", "chessboard").Set("innerHTML", drawBoard(Black))
				}
			case *chesspb.Ping:
				//LogToConsole("[DEBUG] Ping!")
				C.Send(v)
			case *chesspb.OpponentJoined:
				document.Call("getElementById", "newgame").Set("disabled", false)
				LogToConsole("Opponent joined, game is ready to begin!")
			case *chesspb.Player:
				if v.One {
					LogToConsole("Joined as player 1.")
				} else {
					LogToConsole("Joined as player 2.")
				}
			case *chesspb.Team:
				if v.Black {
					LogToConsole("You've been assigned to black.")
					Black = true
					MyTurn = false
				} else {
					LogToConsole("You've been assigned to white.")
					Black = false
					MyTurn = true
				}
				Game = chess.NewChessboard()
				document.Call("getElementById", "chessboard").Set("innerHTML", drawBoard(Black))
				Promotion = nil
				document.Call("getElementById", "blackpromotion").Set("hidden", true)
				document.Call("getElementById", "whitepromotion").Set("hidden", true)
			case *chesspb.Move:
				var check bool
				switch chess.MoveType(v.MoveType) {
				case chess.RegularMove:
					check = Game.DoMove([2]int8{int8(v.Fx), int8(v.Fy)}, [2]int8{int8(v.Tx), int8(v.Ty)})
				case chess.EnPassant:
					check = Game.DoEnPassant([2]int8{int8(v.Fx), int8(v.Fy)}, [2]int8{int8(v.Tx), int8(v.Ty)})
				case chess.CastleLeft:
					check = Game.DoCastle([2]int8{int8(v.Fx), int8(v.Fy)}, true)
				case chess.CastleRight:
					check = Game.DoCastle([2]int8{int8(v.Fx), int8(v.Fy)}, false)
				}
				if check {
					if MyTurn {
						LogToConsole("Your opponent is in check.")
					} else {
						LogToConsole("You are in check.")
					}
				}
				document.Call("getElementById", "chessboard").Set("innerHTML", drawBoard(Black))
				MyTurn = !MyTurn
			case *chesspb.GameComplete:
				switch v.Result {
				case chesspb.GameComplete_Stalemate:
					LogToConsole("Game complete! Stalemate.")
				case chesspb.GameComplete_WhiteWin:
					LogToConsole("Game complete! White wins!")
				case chesspb.GameComplete_BlackWin:
					LogToConsole("Game complete! Black wins!")
				}
				Game = nil
				document.Call("getElementById", "newgame").Set("disabled", true)
			case *chesspb.OpponentLeft:
				LogToConsole("Opponent left, need to rejoin.")
				document.Call("getElementById", "newgame").Set("disabled", true)
				Game = nil
			case *chesspb.Error:
				LogToConsole("Server error: " + v.Msg)
			}
		}
	}()

	return nil
}

func setup() {
	window := js.Global()
	document := window.Get("document")

	// callbacks
	window.Set("connect", js.FuncOf(connect))
	document.Call("getElementById", "connect").Call("setAttribute", "onClick", "connect();")
	window.Set("newgame", js.FuncOf(newGame))
	document.Call("getElementById", "newgame").Call("setAttribute", "onClick", "newgame();")
	window.Set("joingame", js.FuncOf(joinGame))
	document.Call("getElementById", "join").Call("setAttribute", "onClick", "joingame();")

	window.Set("selectpiece", js.FuncOf(selectPiece))
	window.Set("selectPromotion", js.FuncOf(selectPromotion))
}

func main() {
	println("👍")
	LogToConsole("Chess v0.0.0 Loaded.")
	// register functions
	setup()

	<-make(chan bool)
}
