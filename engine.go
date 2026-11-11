package main

const (
	mateScore = 1_000_000
	infinity  = 1_000_000_000
)

// evaluate scores material from White's perspective, in centipawns.
func evaluate(b Board) int {
	values := [7]int{0, 0, 100, 320, 330, 500, 900}
	score := 0
	for _, p := range b.squares {
		value := values[p.pType()]
		if p.pColor() == White {
			score += value
		} else {
			score -= value
		}
	}
	return score
}

// BestMove searches depth plies with minimax and alpha-beta pruning.
// It leaves b unchanged and returns false for depth < 1 or a finished position.
// Repetition history and optional draw claims are not included in the search.
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
		return 0, Move{}, false // Stalemate.
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
