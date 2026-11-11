package main

const (
	mateScore = 1_000_000
	infinity  = 1_000_000_000
)

// positive scores favor white
func evaluate(b Board) int {
	values := [7]int{0, 0, 100, 320, 330, 500, 900}
	score := 0
	for square, p := range b.squares {
		if p == None {
			continue
		}
		if p.pColor() == Black {
			square ^= 56 // flip ranks and keep the file
		}
		value := values[p.pType()] + pieceSquareTables[p.pType()][square]
		if p.pColor() == White {
			score += value
		} else {
			score -= value
		}
	}
	return score
}

// returns no move for an invalid depth or a finished position
// search ignores repetition history and optional draw claims
func BestMove(b Board, depth int) (Move, bool) {
	if depth < 1 {
		return Move{}, false
	}
	_, move, ok := minimax(b, depth, -infinity, infinity, 0)
	return move, ok
}

func minimax(b Board, depth, alpha, beta, ply int) (int, Move, bool) {
	moves := GenerateLegalMoves(b)
	if len(moves) == 0 {
		if kingAttacked(b, b.colorToMove) {
			if b.colorToMove == White {
				return -mateScore + ply, Move{}, false
			}
			return mateScore - ply, Move{}, false
		}
		return 0, Move{}, false
	}
	if b.insufficientMaterial() || b.halfmove >= 150 {
		return 0, Move{}, false
	}
	if depth == 0 {
		return evaluate(b), Move{}, false
	}
	orderMoves(b, moves)

	maximizing := b.colorToMove == White
	bestScore := infinity
	if maximizing {
		bestScore = -infinity
	}
	best := moves[0]
	for _, move := range moves {
		next := b
		next.makeMove(move)
		score, _, _ := minimax(next, depth-1, alpha, beta, ply+1)
		if maximizing {
			if score > bestScore {
				bestScore, best = score, move
			}
			alpha = max(alpha, bestScore)
		} else {
			if score < bestScore {
				bestScore, best = score, move
			}
			beta = min(beta, bestScore)
		}
		if alpha >= beta {
			break
		}
	}
	return bestScore, best, true
}
