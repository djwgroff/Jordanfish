package main

import "fmt"

const startFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

const (
	whiteKingSide byte = 1 << iota
	whiteQueenSide
	blackKingSide
	blackQueenSide
)

type Board struct {
	squares     [64]Piece
	colorToMove Piece
	castling    byte
	enPassant   int // -1 when unavailable
	halfmove    int
	fullmove    int
}

func opposite(color Piece) Piece { return color ^ (White | Black) }

func displayBoard(board Board, perspective Piece) {
	for rank := 0; rank < 8; rank++ {
		r := rank
		if perspective == Black {
			r = 7 - rank
		}
		fmt.Printf("%d ", 8-r)
		for file := 0; file < 8; file++ {
			f := file
			if perspective == Black {
				f = 7 - file
			}
			fmt.Print(PieceUnicodeMap[board.squares[r*8+f]], " ")
		}
		fmt.Println()
	}
	if perspective == Black {
		fmt.Println("  h g f e d c b a")
	} else {
		fmt.Println("  a b c d e f g h")
	}
}

// The complete previous board is a small, reliable undo record, including captures.
func (b *Board) unMakeMove(previous Board) { *b = previous }

// Only positions with identical legal possibilities count toward repetition.
type positionKey struct {
	squares   [64]Piece
	color     Piece
	castling  byte
	enPassant int
}

func (b Board) key() positionKey {
	ep := -1
	if b.enPassant >= 0 {
		for _, m := range GenerateLegalMoves(b) {
			if m.isEnPassant {
				ep = b.enPassant
				break
			}
		}
	}
	return positionKey{b.squares, b.colorToMove, b.castling, ep}
}

// These material combinations cannot produce checkmate, even with cooperation.
func (b Board) insufficientMaterial() bool {
	minors, knights, bishopColor := 0, 0, -1
	for square, p := range b.squares {
		switch p.pType() {
		case None, King:
		case Knight:
			minors++
			knights++
		case Bishop:
			minors++
			color := (square/8 + square%8) % 2
			if bishopColor == -1 {
				bishopColor = color
			} else if bishopColor != color {
				return false
			}
		default:
			return false
		}
	}
	return minors <= 1 || knights == 0
}

func gameResult(b Board, legal []Move, repetitions int) string {
	if len(legal) == 0 {
		if kingAttacked(b, b.colorToMove) {
			if b.colorToMove == White {
				return "Checkmate. Black wins."
			}
			return "Checkmate. White wins."
		}
		return "Draw by stalemate."
	}
	if b.insufficientMaterial() {
		return "Draw by insufficient material."
	}
	if b.halfmove >= 150 {
		return "Draw by the seventy-five-move rule."
	}
	if repetitions >= 5 {
		return "Draw by fivefold repetition."
	}
	return ""
}
