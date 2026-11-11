package main

import (
	"fmt"
	"strconv"
	"strings"
)

func (m Move) String() string {
	s := fmt.Sprintf("%c%c%c%c", 'a'+m.from%8, '8'-m.from/8, 'a'+m.to%8, '8'-m.to/8)
	if m.promotion != None {
		s += map[Piece]string{Queen: "q", Rook: "r", Bishop: "b", Knight: "n"}[m.promotion]
	}
	return s
}

func coordinateToIndex(coordinate string) (int, error) {
	if len(coordinate) != 2 || coordinate[0] < 'a' || coordinate[0] > 'h' || coordinate[1] < '1' || coordinate[1] > '8' {
		return 0, fmt.Errorf("invalid square %q", coordinate)
	}
	return int(coordinate[0]-'a') + int('8'-coordinate[1])*8, nil
}

func parseMove(s string) (Move, error) {
	if len(s) != 4 && len(s) != 5 {
		return Move{}, fmt.Errorf("use a move like e2e4 or a7a8q")
	}
	from, err := coordinateToIndex(s[:2])
	if err != nil {
		return Move{}, err
	}
	to, err := coordinateToIndex(s[2:4])
	if err != nil {
		return Move{}, err
	}
	m := Move{from: from, to: to}
	if len(s) == 5 {
		m.promotion = map[byte]Piece{'q': Queen, 'r': Rook, 'b': Bishop, 'n': Knight}[s[4]]
		if m.promotion == None {
			return Move{}, fmt.Errorf("promotion must be q, r, b, or n")
		}
	}
	return m, nil
}

func boardFromFEN(fen string) (Board, error) {
	b := Board{colorToMove: White, enPassant: -1, fullmove: 1}
	invalid := func() (Board, error) { return Board{}, fmt.Errorf("invalid FEN") }
	parts := strings.Fields(fen)

	if len(parts) != 1 && len(parts) != 6 {
		return invalid()
	}
	ranks := strings.Split(parts[0], "/")
	if len(ranks) != 8 {
		return invalid()
	}
	pieces := map[byte]Piece{'k': King, 'p': Pawn, 'n': Knight, 'b': Bishop, 'r': Rook, 'q': Queen}
	kings := map[Piece]int{}
	for rank, row := range ranks {
		file := 0
		for i := 0; i < len(row); i++ {
			c := row[i]
			if c >= '1' && c <= '8' {
				file += int(c - '0')
			} else {
				color := Black
				if c >= 'A' && c <= 'Z' {
					color = White
					c += 'a' - 'A'
				}
				p, ok := pieces[c]
				if !ok || file >= 8 || (p == Pawn && (rank == 0 || rank == 7)) {
					return invalid()
				}
				b.squares[rank*8+file] = p | color
				if p == King {
					kings[color]++
				}
				file++
			}
			if file > 8 {
				return invalid()
			}
		}
		if file != 8 {
			return invalid()
		}
	}
	if kings[White] != 1 || kings[Black] != 1 {
		return invalid()
	}
	if len(parts) == 1 {
		return b, nil
	}
	switch parts[1] {
	case "w":
	case "b":
		b.colorToMove = Black
	default:
		return invalid()
	}
	if parts[2] != "-" {
		if parts[2] == "" {
			return invalid()
		}
		for _, c := range parts[2] {
			right := map[rune]byte{'K': whiteKingSide, 'Q': whiteQueenSide, 'k': blackKingSide, 'q': blackQueenSide}[c]
			if right == 0 || b.castling&right != 0 {
				return invalid()
			}
			b.castling |= right
		}
	}
	for _, c := range []struct {
		right      byte
		king, rook int
		color      Piece
	}{
		{whiteKingSide, 60, 63, White}, {whiteQueenSide, 60, 56, White},
		{blackKingSide, 4, 7, Black}, {blackQueenSide, 4, 0, Black},
	} {
		if b.castling&c.right != 0 && (b.squares[c.king] != c.color|King || b.squares[c.rook] != c.color|Rook) {
			return invalid()
		}
	}
	if parts[3] != "-" {
		ep, err := coordinateToIndex(parts[3])
		if err != nil {
			return invalid()
		}
		rank, pawn, origin := 2, ep+8, ep-8
		if b.colorToMove == Black {
			rank, pawn, origin = 5, ep-8, ep+8
		}
		if ep/8 != rank || b.squares[ep] != None || b.squares[pawn] != opposite(b.colorToMove)|Pawn || b.squares[origin] != None {
			return invalid()
		}
		b.enPassant = ep
	}
	var err error
	b.halfmove, err = strconv.Atoi(parts[4])
	if err != nil || b.halfmove < 0 {
		return invalid()
	}
	b.fullmove, err = strconv.Atoi(parts[5])
	if err != nil || b.fullmove < 1 {
		return invalid()
	}
	return b, nil
}
