package main

type Piece byte

const None Piece = 0
const King Piece = 1
const Pawn Piece = 2
const Knight Piece = 3
const Bishop Piece = 4
const Rook Piece = 5
const Queen Piece = 6

const White Piece = 8
const Black Piece = 16

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
