// Copyright (c) 2020, The SpeedyChess Contributors. All rights reserved.

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
	CanBeEnPassant                                                                       *[2]int8
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

// IsStalemated returns true if nobody can move.
func (cb *Chessboard) IsStalemated() bool {
	return cb.IsCheckmated(true) && cb.IsCheckmated(false)
}

// IsCheckmated returns true if the colour cannot move anywhere.
func (cb *Chessboard) IsCheckmated(black bool) bool {
	for y, _ := range cb.Board {
		for x, p := range cb.Board[y] {
			if p == 0 || IsBlack(p) != black {
				continue
			}
			moves, enpassant, castleleft, castleright := cb.PossibleMoves(int8(x), int8(y))
			if len(moves) > 0 || len(enpassant) > 0 || castleleft || castleright {
				return false
			}
		}
	}
	return true
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
func (cb *Chessboard) pawnMoves(x, y int8, black bool) (Moves, Threats [][2]int8, EnPassantKill *[2]int8) {
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
	specialThreats := cb.canEnPassant(x, y)
	if specialThreats != nil {
		Threats = append(Threats, *specialThreats)
		EnPassantKill = specialThreats
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

// Returns false if the move would put the player moving in check
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
	black := IsBlack(cb.Board[from[1]][from[0]])
	cb.CanBeEnPassant = nil
	if (black && from[1] == 1 && to[1] == 3 && cb.Board[from[1]][from[0]] == BlackPawn) || (!black && from[1] == 6 && to[1] == 4 && cb.Board[from[1]][from[0]] == WhitePawn) {
		cb.CanBeEnPassant = &to
	}
	if black {
		if cb.Board[from[1]][from[0]] == BlackKing {
			cb.BlackCantCastleLeft = true
			cb.BlackCantCastleRight = true
		} else if cb.Board[from[1]][from[0]] == BlackRook {
			if from[0] == 0 && from[0] == 7 {
				cb.BlackCantCastleLeft = true
			} else if from[0] == 7 && from[0] == 7 {
				cb.BlackCantCastleRight = true
			}
		}
	} else {
		if cb.Board[from[1]][from[0]] == WhiteKing {
			cb.WhiteCantCastleLeft = true
			cb.WhiteCantCastleRight = true
		} else if cb.Board[from[1]][from[0]] == WhiteRook {
			if from[0] == 0 && from[0] == 0 {
				cb.WhiteCantCastleLeft = true
			} else if from[0] == 7 && from[0] == 0 {
				cb.WhiteCantCastleRight = true
			}
		}
	}
	cb.Board[to[1]][to[0]] = cb.Board[from[1]][from[0]]
	cb.Board[from[1]][from[0]] = 0
	if cb.IsCheck(!IsBlack(cb.Board[to[1]][to[0]])) {
		return false
	}
	return true
}

// Returns false if the move would put the player moving in check
func (cb *Chessboard) TestEnPassant(from, to [2]int8) bool {
	board := new(Chessboard)
	*board = *cb
	var modifier int8 // y modifier
	if IsBlack(board.Board[from[1]][from[0]]) {
		modifier = 1
	} else {
		modifier = -1
	}
	board.Board[to[1]][to[0]] = 0
	board.Board[to[1]+modifier][to[0]] = board.Board[from[1]][from[0]]
	board.Board[from[1]][from[0]] = 0
	if board.IsCheck(IsBlack(board.Board[to[1]+modifier][to[0]])) {
		return false
	}
	return true
}

// Performs a move and returns whether a move would result in a check for the opposing player
func (cb *Chessboard) DoEnPassant(from, to [2]int8) bool {
	var modifier int8 // y modifier
	if IsBlack(cb.Board[from[1]][from[0]]) {
		modifier = 1
	} else {
		modifier = -1
	}
	cb.Board[to[1]][to[0]] = 0
	cb.Board[to[1]+modifier][to[0]] = cb.Board[from[1]][from[0]]
	cb.Board[from[1]][from[0]] = 0
	cb.CanBeEnPassant = nil
	if cb.IsCheck(!IsBlack(cb.Board[to[1]+modifier][to[0]])) {
		return false
	}
	return true
}

// Returns false if the move would put the player moving in check
func (cb *Chessboard) TestCastle(from [2]int8, left bool) bool {
	board := new(Chessboard)
	*board = *cb
	black := IsBlack(board.Board[from[1]][from[0]])
	if left {
		board.Board[from[1]][2] = board.Board[from[1]][from[0]]
		board.Board[from[1]][from[0]] = 0
		board.Board[from[1]][3] = board.Board[from[1]][0]
		board.Board[from[1]][0] = 0
	} else {
		board.Board[from[1]][6] = board.Board[from[1]][from[0]]
		board.Board[from[1]][from[0]] = 0
		board.Board[from[1]][5] = board.Board[from[1]][7]
		board.Board[from[1]][7] = 0
	}
	if board.IsCheck(black) {
		return false
	}
	return true
}

// Performs a move and returns whether a move would result in a check for the opposing player
func (cb *Chessboard) DoCastle(from [2]int8, left bool) bool {
	black := IsBlack(cb.Board[from[1]][from[0]])
	if left {
		cb.Board[from[1]][2] = cb.Board[from[1]][from[0]]
		cb.Board[from[1]][from[0]] = 0
		cb.Board[from[1]][3] = cb.Board[from[1]][0]
		cb.Board[from[1]][0] = 0
	} else {
		cb.Board[from[1]][6] = cb.Board[from[1]][from[0]]
		cb.Board[from[1]][from[0]] = 0
		cb.Board[from[1]][5] = cb.Board[from[1]][7]
		cb.Board[from[1]][7] = 0
	}
	if black {
		cb.BlackCantCastleLeft = true
		cb.BlackCantCastleRight = true
	} else {
		cb.WhiteCantCastleLeft = true
		cb.WhiteCantCastleRight = true
	}
	if cb.IsCheck(!black) {
		return false
	}
	return true
}

// PromotePawn promotes a pawn at x, y, returns true on success, and false on failure.
func (cb *Chessboard) PromotePawn(x, y int8, to Piece) bool {
	if y != 0 && y != 7 {
		return false
	}
	piece := cb.Board[y][x]
	black := IsBlack(piece)
	switch to {
	case WhiteRook, WhiteKnight, WhiteQueen, WhiteBishop:
		if !black {
			cb.Board[y][x] = to
			return true
		}
	case BlackRook, BlackKnight, BlackQueen, BlackBishop:
		if black {
			cb.Board[y][x] = to
			return true
		}
	}

	return false
}

type MoveType int

const (
	RegularMove MoveType = iota
	EnPassant
	CastleLeft
	CastleRight
)

// Checks if a piece can move from the from position, to the to position.
func (cb *Chessboard) IsLegal(from, to [2]int8, movet MoveType) bool {
	moves, enpassantkill, castleleft, castleright := cb.PossibleMoves(from[0], from[1])
	switch movet {
	case RegularMove:
		for _, move := range moves {
			if to == move {
				return true
			}
		}
	case EnPassant:
		for _, move := range enpassantkill {
			if to == move {
				return true
			}
		}
	case CastleLeft:
		return castleleft
	case CastleRight:
		return castleright
	}
	return false
}

// canCastle checks if the king at x/y can castle. Returns 2 bools, first is `CanCastleLeft`, second is `CanCastleRight`.
func (cb *Chessboard) canCastle(x, y int8) (CanCastleLeft, CanCastleRight bool) {
	if cb.Board[y][x] != BlackKing && cb.Board[y][x] != WhiteKing {
		return
	}
	black := IsBlack(cb.Board[y][x])
	if ((black && !cb.BlackCantCastleLeft) || (!black && !cb.WhiteCantCastleLeft)) && cb.Board[y][x-1] == 0 && cb.Board[y][x-2] == 0 && cb.Board[y][x-3] == 0 {
		CanCastleLeft = true
	}
	if ((black && !cb.BlackCantCastleRight) || (!black && !cb.WhiteCantCastleRight)) && cb.Board[y][x+1] == 0 && cb.Board[y][x+2] == 0 {
		CanCastleRight = true
	}
	return
}

// Should return all legal moves by a piece at x, y.
func (cb *Chessboard) PossibleMoves(x, y int8) (OutMoves, OutEnPassantKill [][2]int8, CanCastleLeft, CanCastleRight bool) {
	var (
		Moves         [][2]int8
		EnPassantKill *[2]int8
	)
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
		var pawnMoves [][2]int8
		pawnMoves, _, EnPassantKill = cb.pawnMoves(x, y, true)
		Moves = append(Moves, pawnMoves...)
	case WhitePawn:
		var pawnMoves [][2]int8
		pawnMoves, _, EnPassantKill = cb.pawnMoves(x, y, false)
		Moves = append(Moves, pawnMoves...)
	case BlackKing:
		Moves = append(Moves, cb.kingMoves(x, y, true)...)
		CanCastleLeft, CanCastleRight = cb.canCastle(x, y)
	case WhiteKing:
		Moves = append(Moves, cb.kingMoves(x, y, false)...)
		CanCastleLeft, CanCastleRight = cb.canCastle(x, y)
	}

	// Test if they're *really* possible.
	from := [2]int8{x, y}
	for _, move := range Moves {
		if cb.TestMove(from, move) {
			OutMoves = append(OutMoves, move)
		}
	}
	if EnPassantKill != nil {
		move := *EnPassantKill
		if cb.TestEnPassant(from, move) {
			OutEnPassantKill = [][2]int8{move}
		}
	}
	if CanCastleLeft {
		CanCastleLeft = cb.TestCastle(from, true)
	}
	if CanCastleRight {
		CanCastleRight = cb.TestCastle(from, false)
	}
	return
}

// canEnPassant returns a piece the pawn in x, y could eliminate via en passant or nil.
func (cb *Chessboard) canEnPassant(x, y int8) *[2]int8 {
	if cb.Board[y][x] != BlackPawn && cb.Board[y][x] != WhitePawn || cb.CanBeEnPassant == nil || y < 2 || y > 5 {
		return nil
	}
	var modifier int8 // y modifier
	if IsBlack(cb.Board[y][x]) {
		modifier = 1
	} else {
		modifier = -1
	}
	var out *[2]int8
	black := IsBlack(cb.Board[y][x])
	move := *cb.CanBeEnPassant
	if y == move[1] && IsBlack(cb.Board[move[1]][move[0]]) != black && cb.Board[move[1]+modifier][move[0]] == 0 {
		if x > 0 && x-1 == move[0] {
			out = &[2]int8{move[0], move[1]}
		}
		if x < 7 && x+1 == move[0] {
			out = &[2]int8{move[0], move[1]}
		}
	}
	return out
}

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
		_, pawnThreats, _ := cb.pawnMoves(x, y, true)
		Moves = append(Moves, pawnThreats...)
	case WhitePawn:
		_, pawnThreats, _ := cb.pawnMoves(x, y, false)
		Moves = append(Moves, pawnThreats...)
	case BlackKing, WhiteKing:
		Moves = append(Moves, cb.kingMoves(x, y, true)...)
		Moves = append(Moves, cb.kingMoves(x, y, false)...)
	}
	return
}
