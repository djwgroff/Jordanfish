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
	table := make(transpositionTable)
	var best Move
	for currentDepth := 1; currentDepth <= depth; currentDepth++ {
		_, move, ok := search(b, currentDepth, -infinity, infinity, 0, table)
		if !ok {
			return Move{}, false
		}
		best = move
	}
	return best, true
}

func minimax(b Board, depth, alpha, beta, ply int) (int, Move, bool) {
	return search(b, depth, alpha, beta, ply, make(transpositionTable))
}

func search(b Board, depth, alpha, beta, ply int, table transpositionTable) (int, Move, bool) {
	if depth == 0 {
		return quiescence(b, alpha, beta, ply, 0), Move{}, false
	}
	key := tableKey(b)
	entry, cached := table[key]
	if cached && entry.depth >= depth {
		score := scoreFromTable(entry.score, ply)
		if entry.bound == exactScore || entry.bound == lowerBound && score >= beta || entry.bound == upperBound && score <= alpha {
			return score, entry.move, true
		}
	}
	originalAlpha, originalBeta := alpha, beta
	moves := GenerateLegalMoves(b)
	if score, finished := terminalScore(b, moves, ply); finished {
		return score, Move{}, false
	}
	orderMoves(b, moves)
	if cached {
		for i, move := range moves {
			if move == entry.move {
				moves[0], moves[i] = moves[i], moves[0]
				break
			}
		}
	}

	maximizing := b.colorToMove == White
	bestScore := infinity
	if maximizing {
		bestScore = -infinity
	}
	best := moves[0]
	for _, move := range moves {
		next := b
		next.makeMove(move)
		score, _, _ := search(next, depth-1, alpha, beta, ply+1, table)
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
	bound := exactScore
	if bestScore <= originalAlpha {
		bound = upperBound
	} else if bestScore >= originalBeta {
		bound = lowerBound
	}
	table.store(key, tableEntry{depth, scoreToTable(bestScore, ply), bound, best})
	return bestScore, best, true
}

func terminalScore(b Board, moves []Move, ply int) (int, bool) {
	if len(moves) == 0 {
		if kingAttacked(b, b.colorToMove) {
			if b.colorToMove == White {
				return -mateScore + ply, true
			}
			return mateScore - ply, true
		}
		return 0, true
	}
	if b.insufficientMaterial() || b.halfmove >= 150 {
		return 0, true
	}
	return 0, false
}

const maxQuiescencePly = 16

func quiescence(b Board, alpha, beta, ply, extraPly int) int {
	moves := GenerateLegalMoves(b)
	if score, finished := terminalScore(b, moves, ply); finished {
		return score
	}
	standPat := evaluate(b)
	// cap extensions so checking sequences cannot run indefinitely
	if extraPly >= maxQuiescencePly {
		return standPat
	}

	maximizing := b.colorToMove == White
	inCheck := kingAttacked(b, b.colorToMove)
	bestScore := standPat
	if inCheck {
		// being in check means we have to search a reply instead of standing pat
		bestScore = infinity
		if maximizing {
			bestScore = -infinity
		}
	} else if maximizing {
		if standPat >= beta {
			return standPat
		}
		alpha = max(alpha, standPat)
	} else {
		if standPat <= alpha {
			return standPat
		}
		beta = min(beta, standPat)
	}
	orderMoves(b, moves)
	for _, move := range moves {
		if !inCheck && b.squares[move.to] == None && !move.isEnPassant && move.promotion == None {
			continue
		}
		next := b
		next.makeMove(move)
		score := quiescence(next, alpha, beta, ply+1, extraPly+1)
		if maximizing {
			bestScore = max(bestScore, score)
			alpha = max(alpha, bestScore)
		} else {
			bestScore = min(bestScore, score)
			beta = min(beta, bestScore)
		}
		if alpha >= beta {
			break
		}
	}
	return bestScore
}
