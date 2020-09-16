package chess

type Piece rune

const (
	BlackRook   Piece = '♜'
	WhiteRook   Piece = '♖'
	BlackKnight Piece = '♞'
	WhiteKnight Piece = '♘'
	BlackKing   Piece = '♚'
	WhiteKing   Piece = '♔'
	BlackQueen  Piece = '♛'
	WhiteQueen  Piece = '♕'
	BlackBishop Piece = '♝'
	WhiteBishop Piece = '♗'
	BlackPawn   Piece = '♟'
	WhitePawn   Piece = '♙'
)

func IsBlack(p Piece) bool {
	switch p {
	default:
		return false
	case BlackRook, BlackKnight, BlackKing, BlackQueen, BlackBishop, BlackPawn:
		return true
	}
}

type Chessboard struct {
	Board                                                                                [8][8]Piece
	WhiteCantCastleLeft, WhiteCantCastleRight, BlackCantCastleLeft, BlackCantCastleRight bool
	CanBeEnPassant                                                                       [][2]int8
}

func NewChessboard() *Chessboard {
	return &Chessboard{Board: [8][8]Piece{
		[8]Piece{'♜', '♞', '♝', '♛', '♚', '♝', '♞', '♜'},
		[8]Piece{'♟', '♟', '♟', '♟', '♟', '♟', '♟', '♟'},
		[8]Piece{0, 0, 0, 0, 0, 0, 0, 0},
		[8]Piece{0, 0, 0, 0, 0, 0, 0, 0},
		[8]Piece{0, 0, 0, 0, 0, 0, 0, 0},
		[8]Piece{0, 0, 0, 0, 0, 0, 0, 0},
		[8]Piece{'♙', '♙', '♙', '♙', '♙', '♙', '♙', '♙'},
		[8]Piece{'♖', '♘', '♗', '♕', '♔', '♗', '♘', '♖'},
	}}
}

// Threat returns all threatened spaces by a particular colour.
func (cb *Chessboard) Threat(black bool) (threatBoard [8][8]bool) {
	var threatMoves [][2]int8
	for y, _ := range cb.Board {
		for x, p := range cb.Board[y] {
			if p == 0 || IsBlack(p) != black {
				continue
			}
			if possibleThreats := cb.PossibleThreats(int8(x), int8(y)); possibleThreats != nil {
				threatMoves = append(threatMoves, possibleThreats...)
			}
		}
	}
	for _, pos := range threatMoves {
		threatBoard[pos[1]][pos[0]] = true
	}
	return
}

// move returns a possible move (nil if none), and if it hit a piece
func (cb *Chessboard) move(x, y int8, black bool) (*[2]int8, bool) {
	if x < 0 || x > 7 || y < 0 || y > 7 {
		return nil, false
	}
	if cb.Board[y][x] != 0 {
		if IsBlack(cb.Board[y][x]) != black {
			return &[2]int8{x, y}, true
		}
		return nil, true
	}
	return &[2]int8{x, y}, false
}

// pawnMoves returns both moves and threatened spaces
// TODO en passant
func (cb *Chessboard) pawnMoves(x, y int8, black bool) (Moves, Threats [][2]int8) {
	if black {
		// regular movement
		move, hit := cb.move(x, y+1, black)
		if move != nil && !hit {
			Moves = append(Moves, [2]int8{move[0], move[1]})
			if y == 1 { // double-move
				move, hit = cb.move(x, y+2, black)
				if move != nil && !hit {
					Moves = append(Moves, [2]int8{move[0], move[1]})
				}
			}
		}
		// kill moves
		move, hit = cb.move(x+1, y+1, black)
		if move != nil {
			if hit {
				Moves = append(Moves, [2]int8{move[0], move[1]})
			}
			Threats = append(Threats, [2]int8{move[0], move[1]})
		} else if hit {
			Threats = append(Threats, [2]int8{x + 1, y + 1})
		}
		move, hit = cb.move(x-1, y+1, black)
		if move != nil && hit {
			if hit {
				Moves = append(Moves, [2]int8{move[0], move[1]})
			}
			Threats = append(Threats, [2]int8{move[0], move[1]})
		} else if hit {
			Threats = append(Threats, [2]int8{x - 1, y + 1})
		}
	} else {
		move, hit := cb.move(x, y-1, black)
		if move != nil && !hit {
			Moves = append(Moves, [2]int8{move[0], move[1]})
			if y == 6 { // double-move
				move, hit = cb.move(x, y-2, black)
				if move != nil && !hit {
					Moves = append(Moves, [2]int8{move[0], move[1]})
				}
			}
		}
		// kill moves
		move, hit = cb.move(x+1, y-1, black)
		if move != nil {
			if hit {
				Moves = append(Moves, [2]int8{move[0], move[1]})
			}
			Threats = append(Threats, [2]int8{move[0], move[1]})
		} else if hit {
			Threats = append(Threats, [2]int8{x + 1, y - 1})
		}
		move, hit = cb.move(x-1, y-1, black)
		if move != nil && hit {
			if hit {
				Moves = append(Moves, [2]int8{move[0], move[1]})
			}
			Threats = append(Threats, [2]int8{move[0], move[1]})
		} else if hit {
			Threats = append(Threats, [2]int8{x - 1, y - 1})
		}
	}
	return
}

func (cb *Chessboard) kingMoves(x, y int8, black bool) (Moves [][2]int8) {
	// up
	move, _ := cb.move(x, y-1, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// up-right
	move, _ = cb.move(x+1, y-1, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// right
	move, _ = cb.move(x+1, y, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// down-right
	move, _ = cb.move(x+1, y+1, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// down
	move, _ = cb.move(x, y+1, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// down-left
	move, _ = cb.move(x-1, y+1, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// left
	move, _ = cb.move(x-1, y, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// up-left
	move, _ = cb.move(x-1, y-1, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	return
}

func (cb *Chessboard) knightMoves(x, y int8, black bool) (Moves [][2]int8) {
	// up-left
	move, _ := cb.move(x-1, y-2, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// up-right
	move, _ = cb.move(x+1, y-2, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// right-up
	move, _ = cb.move(x+2, y-1, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// right-down
	move, _ = cb.move(x+2, y+1, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// down-right
	move, _ = cb.move(x+1, y+2, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// down-left
	move, _ = cb.move(x-1, y+2, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// left-down
	move, _ = cb.move(x-2, y+1, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	// left-up
	move, _ = cb.move(x-2, y-1, black)
	if move != nil {
		Moves = append(Moves, [2]int8{move[0], move[1]})
	}
	return
}

func (cb *Chessboard) bishopMoves(x, y int8, black bool) (Moves [][2]int8) {
	// up-left
	for ty, tx := y-1, x-1; tx >= 0 && ty >= 0; tx, ty = tx-1, ty-1 {
		move, hit := cb.move(tx, ty, black)
		if move != nil {
			Moves = append(Moves, [2]int8{move[0], move[1]})
		}
		if hit {
			break
		}
	}
	// up-right
	for ty, tx := y-1, x+1; tx < 8 && ty >= 0; tx, ty = tx+1, ty-1 {
		move, hit := cb.move(tx, ty, black)
		if move != nil {
			Moves = append(Moves, [2]int8{move[0], move[1]})
		}
		if hit {
			break
		}
	}
	// down-right
	for ty, tx := y+1, x-1; tx >= 0 && ty < 8; tx, ty = tx-1, ty+1 {
		move, hit := cb.move(tx, ty, black)
		if move != nil {
			Moves = append(Moves, [2]int8{move[0], move[1]})
		}
		if hit {
			break
		}
	}
	// down-left
	for ty, tx := y+1, x+1; tx < 8 && ty <= 8; tx, ty = tx+1, ty+1 {
		move, hit := cb.move(tx, ty, black)
		if move != nil {
			Moves = append(Moves, [2]int8{move[0], move[1]})
		}
		if hit {
			break
		}
	}
	return
}

func (cb *Chessboard) rookMoves(x, y int8, black bool) (Moves [][2]int8) {
	// up
	for ty := y - 1; ty >= 0; ty-- {
		move, hit := cb.move(x, ty, black)
		if move != nil {
			Moves = append(Moves, [2]int8{move[0], move[1]})
		}
		if hit {
			break
		}
	}
	// down
	for ty := y + 1; ty < 8; ty++ {
		move, hit := cb.move(x, ty, black)
		if move != nil {
			Moves = append(Moves, [2]int8{move[0], move[1]})
		}
		if hit {
			break
		}
	}
	// left
	for tx := x - 1; tx >= 0; tx-- {
		move, hit := cb.move(tx, y, black)
		if move != nil {
			Moves = append(Moves, [2]int8{move[0], move[1]})
		}
		if hit {
			break
		}
	}
	// right
	for tx := x + 1; tx < 8; tx++ {
		move, hit := cb.move(tx, y, black)
		if move != nil {
			Moves = append(Moves, [2]int8{move[0], move[1]})
		}
		if hit {
			break
		}
	}
	return
}

// Returns whether a particular colour is in check.
func (cb *Chessboard) IsCheck(black bool) bool {
	var kx, ky int
	for y, row := range cb.Board {
		for x, _ := range row {
			if (black && cb.Board[y][x] == BlackKing) || (!black && cb.Board[y][x] == WhiteKing) {
				kx, ky = x, y
				break
			}
		}
	}
	return cb.Threat(!black)[ky][kx]
}

// Returns whether a move would result in a check for the player moving
func (cb *Chessboard) TestMove(from, to [2]int8) bool {
	board := new(Chessboard)
	*board = *cb
	board.Board[to[1]][to[0]] = board.Board[from[1]][from[0]]
	board.Board[from[1]][from[0]] = 0
	if board.IsCheck(IsBlack(board.Board[to[1]][to[0]])) {
		return false
	}
	return true
}

// Performs a move and returns whether a move would result in a check for the opposing player
func (cb *Chessboard) DoMove(from, to [2]int8) bool {
	cb.Board[to[1]][to[0]] = cb.Board[from[1]][from[0]]
	cb.Board[from[1]][from[0]] = 0
	if cb.IsCheck(!IsBlack(cb.Board[to[1]][to[0]])) {
		return false
	}
	return true
}

// Checks if a piece can move from the from position, to the to position.
func (cb *Chessboard) IsLegal(from, to [2]int8) bool {
	moves := cb.PossibleMoves(from[0], from[1])
	for _, move := range moves {
		if to == move {
			return true
		}
	}
	return false
}

// Should return all legal moves by a piece at x, y.
func (cb *Chessboard) PossibleMoves(x, y int8) (OutMoves [][2]int8) {
	var Moves [][2]int8
	switch P := cb.Board[y][x]; P {
	case BlackRook:
		Moves = append(Moves, cb.rookMoves(x, y, true)...)
	case WhiteRook:
		Moves = append(Moves, cb.rookMoves(x, y, false)...)
	case BlackKnight:
		Moves = append(Moves, cb.knightMoves(x, y, true)...)
	case WhiteKnight:
		Moves = append(Moves, cb.knightMoves(x, y, false)...)
	case BlackBishop:
		Moves = append(Moves, cb.bishopMoves(x, y, true)...)
	case WhiteBishop:
		Moves = append(Moves, cb.bishopMoves(x, y, false)...)
	case BlackQueen:
		Moves = append(Moves, cb.rookMoves(x, y, true)...)
		Moves = append(Moves, cb.bishopMoves(x, y, true)...)
	case WhiteQueen:
		Moves = append(Moves, cb.rookMoves(x, y, false)...)
		Moves = append(Moves, cb.bishopMoves(x, y, false)...)
	case BlackPawn:
		pawnMoves, _ := cb.pawnMoves(x, y, true)
		Moves = append(Moves, pawnMoves...)
	case WhitePawn:
		pawnMoves, _ := cb.pawnMoves(x, y, false)
		Moves = append(Moves, pawnMoves...)
	case BlackKing:
		Moves = append(Moves, cb.kingMoves(x, y, true)...)
	case WhiteKing:
		Moves = append(Moves, cb.kingMoves(x, y, false)...)
	}
	from := [2]int8{x, y}
	for _, move := range Moves {
		if cb.TestMove(from, move) {
			OutMoves = append(OutMoves, move)
		}
	}
	return
}

/*
// CanEnPassant returns the possible spaces the pawn could jump to
func (cb *Chessboard) CanEnPassant(x, y int8) [][2]int8 {
	if cb.Board[y][x] != BlackPawn && cb.Board[y][x] != WhitePawn || len(cb.CanBeEnPassant) == 0 || y < 2 || y > 5 {
		return nil
	}
	out := make([][2]int8, 0)
	black := IsBlack(cb.Board[y][x])
	for _, move := range cb.CanEnPassant {
		if y == move[1] && IsBlack(cb.Board[move[1]][move[0]]) != black {
			if x > 0 && x-1 == move[0] {
				out = append(out, [2]int8{move[0], move[1]})
			}
			if x < 7 && x+1 == move[0] {
				out = append(out, [2]int8{move[0], move[1]})
			}
		}
	}
	return out
}
*/
// PossibleThreats returns all the possibly threatened spaces, can contain duplicates
func (cb *Chessboard) PossibleThreats(x, y int8) (Moves [][2]int8) {
	switch P := cb.Board[y][x]; P {
	case BlackRook, WhiteRook:
		Moves = append(Moves, cb.rookMoves(x, y, true)...)
		Moves = append(Moves, cb.rookMoves(x, y, false)...)
	case BlackKnight, WhiteKnight:
		Moves = append(Moves, cb.knightMoves(x, y, true)...)
		Moves = append(Moves, cb.knightMoves(x, y, false)...)
	case BlackBishop, WhiteBishop:
		Moves = append(Moves, cb.bishopMoves(x, y, true)...)
		Moves = append(Moves, cb.bishopMoves(x, y, false)...)
	case BlackQueen, WhiteQueen:
		Moves = append(Moves, cb.rookMoves(x, y, true)...)
		Moves = append(Moves, cb.bishopMoves(x, y, true)...)
		Moves = append(Moves, cb.rookMoves(x, y, false)...)
		Moves = append(Moves, cb.bishopMoves(x, y, false)...)
	case BlackPawn:
		_, pawnThreats := cb.pawnMoves(x, y, true)
		Moves = append(Moves, pawnThreats...)
	case WhitePawn:
		_, pawnThreats := cb.pawnMoves(x, y, false)
		Moves = append(Moves, pawnThreats...)
	case BlackKing, WhiteKing:
		Moves = append(Moves, cb.kingMoves(x, y, true)...)
		Moves = append(Moves, cb.kingMoves(x, y, false)...)
	}
	return
}
