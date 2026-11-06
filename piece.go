package main

type Piece byte

const None Piece = 0   // 00000000
const King Piece = 1   // 00000001
const Pawn Piece = 2   // 00000010
const Knight Piece = 3 // 00000011
const Bishop Piece = 4 // 00000100
const Rook Piece = 5   // 00000101
const Queen Piece = 6  // 00000110

const White Piece = 8  // 00001000	>> 8 / 8 % 2 * 2 - 1
const Black Piece = 16 // 00010000	>> 16 / 8 % 2 * 2 - 1

var PieceUnicodeMap = map[Piece]string{
	None: "·",

	White | King:   "♔",
	White | Pawn:   "♙",
	White | Knight: "♘",
	White | Bishop: "♗",
	White | Rook:   "♖",
	White | Queen:  "♕",

	Black | King:   "♚",
	Black | Pawn:   "♟",
	Black | Knight: "♞",
	Black | Bishop: "♝",
	Black | Rook:   "♜",
	Black | Queen:  "♛",
}

func (p Piece) pColor() Piece {
	return p & (White | Black)
}

func (p Piece) pType() Piece {
	return p & (Piece(0b00000111))
}

func (p Piece) isSlidingPiece() bool {
	pieceType := p & (Piece(0b00000111))
	return pieceType == Bishop || pieceType == Rook || pieceType == Queen
}
