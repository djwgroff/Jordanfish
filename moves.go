package main

type Move struct {
	from, to    int
	promotion   Piece
	isEnPassant bool
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

var directions = [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
var knightSteps = [][2]int{{-2, -1}, {-2, 1}, {2, -1}, {2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}}

func onBoard(rank, file int) bool { return rank >= 0 && rank < 8 && file >= 0 && file < 8 }

func (b *Board) makeMove(m Move) Board {
	previous := *b
	p, captured := b.squares[m.from], b.squares[m.to]
	b.squares[m.to], b.squares[m.from] = p, None
	if m.isEnPassant {
		b.squares[m.from/8*8+m.to%8] = None
	}
	if m.promotion != None {
		b.squares[m.to] = p.pColor() | m.promotion
	}
	if p.pType() == King {
		if abs(m.to-m.from) == 2 {
			rookFrom, rookTo := m.from+3, m.from+1
			if m.to < m.from {
				rookFrom, rookTo = m.from-4, m.from-1
			}
			b.squares[rookTo], b.squares[rookFrom] = b.squares[rookFrom], None
		}
		if p.pColor() == White {
			b.castling &^= whiteKingSide | whiteQueenSide
		} else {
			b.castling &^= blackKingSide | blackQueenSide
		}
	}

	for _, c := range []struct {
		square int
		right  byte
	}{{63, whiteKingSide}, {56, whiteQueenSide}, {7, blackKingSide}, {0, blackQueenSide}} {
		if m.from == c.square || m.to == c.square {
			b.castling &^= c.right
		}
	}
	b.enPassant = -1
	if p.pType() == Pawn && abs(m.to-m.from) == 16 {
		b.enPassant = (m.from + m.to) / 2
	}
	b.halfmove++
	if p.pType() == Pawn || captured != None || m.isEnPassant {
		b.halfmove = 0
	}
	if b.colorToMove == Black {
		b.fullmove++
	}
	b.colorToMove = opposite(b.colorToMove)
	return previous
}

func GenerateLegalMoves(b Board) []Move {
	var legal []Move
	for _, m := range GenerateMoves(b) {
		next := b
		next.makeMove(m)
		if !kingAttacked(next, b.colorToMove) {
			legal = append(legal, m)
		}
	}
	return legal
}

func kingAttacked(b Board, color Piece) bool {
	for square, p := range b.squares {
		if p == color|King {
			return squareAttacked(b, square, opposite(color))
		}
	}
	return true
}

// pinned pieces still attack squares even when they cannot move
func squareAttacked(b Board, square int, by Piece) bool {
	r, f := square/8, square%8
	for from, p := range b.squares {
		if p == None || p.pColor() != by {
			continue
		}
		dr, df := r-from/8, f-from%8
		switch p.pType() {
		case Pawn:
			step := -1
			if by == Black {
				step = 1
			}
			if dr == step && abs(df) == 1 {
				return true
			}
		case Knight:
			if abs(dr)*abs(df) == 2 {
				return true
			}
		case King:
			if abs(dr) <= 1 && abs(df) <= 1 {
				return true
			}
		default:
			diagonal := abs(dr) == abs(df) && dr != 0
			straight := (dr == 0) != (df == 0)
			if !((diagonal && (p.pType() == Bishop || p.pType() == Queen)) || (straight && (p.pType() == Rook || p.pType() == Queen))) {
				continue
			}
			rstep, fstep := 0, 0
			if dr != 0 {
				rstep = dr / abs(dr)
			}
			if df != 0 {
				fstep = df / abs(df)
			}
			clear := true
			for rr, ff := from/8+rstep, from%8+fstep; rr != r || ff != f; rr, ff = rr+rstep, ff+fstep {
				if b.squares[rr*8+ff] != None {
					clear = false
					break
				}
			}
			if clear {
				return true
			}
		}
	}
	return false
}

func GenerateMoves(b Board) []Move {
	var moves []Move
	for from, p := range b.squares {
		if p == None || p.pColor() != b.colorToMove {
			continue
		}
		r, f := from/8, from%8
		add := func(to int, ep bool) {
			if p.pType() == Pawn && (to/8 == 0 || to/8 == 7) {
				for _, promotion := range []Piece{Queen, Rook, Bishop, Knight} {
					moves = append(moves, Move{from, to, promotion, ep})
				}
			} else {
				moves = append(moves, Move{from: from, to: to, isEnPassant: ep})
			}
		}
		if p.pType() == Pawn {
			step, start := -1, 6
			if p.pColor() == Black {
				step, start = 1, 1
			}
			if !onBoard(r+step, f) {
				continue
			}
			to := (r+step)*8 + f
			if b.squares[to] == None {
				add(to, false)
				if r == start && b.squares[to+step*8] == None {
					add(to+step*8, false)
				}
			}
			for _, df := range []int{-1, 1} {
				if !onBoard(r+step, f+df) {
					continue
				}
				target := to + df
				q := b.squares[target]
				if q != None && q.pColor() != p.pColor() && q.pType() != King {
					add(target, false)
				}
				if target == b.enPassant && q == None && b.squares[r*8+f+df] == opposite(p.pColor())|Pawn {
					add(target, true)
				}
			}
			continue
		}
		steps := directions
		if p.pType() == Knight {
			steps = knightSteps
		}
		for i, d := range steps {
			if p.pType() == Bishop && i < 4 || p.pType() == Rook && i >= 4 {
				continue
			}
			for rr, ff := r+d[0], f+d[1]; onBoard(rr, ff); rr, ff = rr+d[0], ff+d[1] {
				to := rr*8 + ff
				q := b.squares[to]
				if q.pColor() == p.pColor() || q.pType() == King {
					break
				}
				add(to, false)
				if q != None || !p.isSlidingPiece() {
					break
				}
			}
		}
		if p.pType() == King {
			moves = append(moves, castleMoves(b, from)...)
		}
	}
	return moves
}

func castleMoves(b Board, from int) []Move {
	home, kingSide, queenSide := 60, whiteKingSide, whiteQueenSide
	if b.colorToMove == Black {
		home, kingSide, queenSide = 4, blackKingSide, blackQueenSide
	}
	if from != home || kingAttacked(b, b.colorToMove) {
		return nil
	}
	var moves []Move
	for _, c := range []struct {
		right      byte
		step, rook int
	}{{kingSide, 1, home + 3}, {queenSide, -1, home - 4}} {
		if b.castling&c.right == 0 || b.squares[c.rook] != b.colorToMove|Rook {
			continue
		}
		clear := true
		for sq := home + c.step; sq != c.rook; sq += c.step {
			if b.squares[sq] != None {
				clear = false
				break
			}
		}
		if !clear {
			continue
		}
		// move the king off its starting square before checking the transit square
		transit := b
		transit.makeMove(Move{from: home, to: home + c.step})
		if !kingAttacked(transit, b.colorToMove) {
			moves = append(moves, Move{from: home, to: home + 2*c.step})
		}
	}
	return moves
}
