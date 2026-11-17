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
	return generateLegalMoves(b, false, nil)
}

// captures and promotions without quiet moves
func GenerateLegalTacticalMoves(b Board) []Move {
	return generateLegalMoves(b, true, nil)
}

func findKing(b Board, color Piece) int {
	for square, p := range b.squares {
		if p == color|King {
			return square
		}
	}
	return -1
}

func legalMove(b Board, m Move, kingSquare int) bool {
	b.makeMove(m)
	if m.from == kingSquare {
		kingSquare = m.to
	}
	return !squareAttacked(b, kingSquare, b.colorToMove)
}

func generateLegalMoves(b Board, tacticalOnly bool, buffer []Move) []Move {
	kingSquare := findKing(b, b.colorToMove)
	if kingSquare < 0 {
		return buffer[:0]
	}
	moves := collectMoves(b, tacticalOnly, buffer)
	legal := moves[:0]
	for _, m := range moves {
		if legalMove(b, m, kingSquare) {
			legal = append(legal, m)
		}
	}
	return legal
}

func hasLegalMove(b Board) bool {
	kingSquare := findKing(b, b.colorToMove)
	if kingSquare < 0 {
		return false
	}
	found := false
	visitMoves(b, false, func(m Move) bool {
		found = legalMove(b, m, kingSquare)
		return found
	})
	return found
}

func kingAttacked(b Board, color Piece) bool {
	square := findKing(b, color)
	if square < 0 {
		return true
	}
	return squareAttacked(b, square, opposite(color))
}

// pinned pieces still attack squares even when they cannot move
func squareAttacked(b Board, square int, by Piece) bool {
	r, f := square/8, square%8
	pawnRank := r + 1
	if by == Black {
		pawnRank = r - 1
	}
	for _, df := range []int{-1, 1} {
		if onBoard(pawnRank, f+df) && b.squares[pawnRank*8+f+df] == by|Pawn {
			return true
		}
	}
	for _, d := range knightSteps {
		rr, ff := r+d[0], f+d[1]
		if onBoard(rr, ff) && b.squares[rr*8+ff] == by|Knight {
			return true
		}
	}
	for i, d := range directions {
		for rr, ff := r+d[0], f+d[1]; onBoard(rr, ff); rr, ff = rr+d[0], ff+d[1] {
			p := b.squares[rr*8+ff]
			if p == None {
				continue
			}
			if p.pColor() == by && (p.pType() == Queen ||
				i < 4 && p.pType() == Rook || i >= 4 && p.pType() == Bishop ||
				p.pType() == King && abs(rr-r) <= 1 && abs(ff-f) <= 1) {
				return true
			}
			break
		}
	}
	return false
}

func GenerateMoves(b Board) []Move {
	return collectMoves(b, false, nil)
}

func collectMoves(b Board, tacticalOnly bool, buffer []Move) []Move {
	moves := buffer[:0]
	visitMoves(b, tacticalOnly, func(m Move) bool {
		moves = append(moves, m)
		return false
	})
	return moves
}

// returning true from visit stops generation early
func visitMoves(b Board, tacticalOnly bool, visit func(Move) bool) {
	for from, p := range b.squares {
		if p == None || p.pColor() != b.colorToMove {
			continue
		}
		r, f := from/8, from%8
		add := func(to int, ep bool) bool {
			if p.pType() == Pawn && (to/8 == 0 || to/8 == 7) {
				for _, promotion := range []Piece{Queen, Rook, Bishop, Knight} {
					if visit(Move{from, to, promotion, ep}) {
						return true
					}
				}
				return false
			}
			return visit(Move{from: from, to: to, isEnPassant: ep})
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
			if b.squares[to] == None && (!tacticalOnly || to/8 == 0 || to/8 == 7) {
				if add(to, false) {
					return
				}
				if !tacticalOnly && r == start && b.squares[to+step*8] == None {
					if add(to+step*8, false) {
						return
					}
				}
			}
			for _, df := range []int{-1, 1} {
				if !onBoard(r+step, f+df) {
					continue
				}
				target := to + df
				q := b.squares[target]
				if q != None && q.pColor() != p.pColor() && q.pType() != King {
					if add(target, false) {
						return
					}
				}
				if target == b.enPassant && q == None && b.squares[r*8+f+df] == opposite(p.pColor())|Pawn {
					if add(target, true) {
						return
					}
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
				if !tacticalOnly || q != None {
					if add(to, false) {
						return
					}
				}
				if q != None || !p.isSlidingPiece() {
					break
				}
			}
		}
		if !tacticalOnly && p.pType() == King {
			for _, m := range castleMoves(b, from) {
				if visit(m) {
					return
				}
			}
		}
	}
}

func castleMoves(b Board, from int) []Move {
	home, kingSide, queenSide := 60, whiteKingSide, whiteQueenSide
	if b.colorToMove == Black {
		home, kingSide, queenSide = 4, blackKingSide, blackQueenSide
	}
	if b.castling&(kingSide|queenSide) == 0 || from != home || squareAttacked(b, home, opposite(b.colorToMove)) {
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
		if !squareAttacked(transit, home+c.step, opposite(b.colorToMove)) {
			moves = append(moves, Move{from: home, to: home + 2*c.step})
		}
	}
	return moves
}
