package main

import "sort"

const historyLimit = 16384

type moveHeuristics struct {
	killers [128][2]Move
	history [2][64][64]int
}

func isQuietMove(b Board, m Move) bool {
	return b.squares[m.to] == None && !m.isEnPassant && m.promotion == None
}

func (h *moveHeuristics) recordCutoff(b Board, m Move, depth, ply int) {
	if !isQuietMove(b, m) {
		return
	}
	if ply >= 0 && ply < len(h.killers) && h.killers[ply][0] != m {
		h.killers[ply][1] = h.killers[ply][0]
		h.killers[ply][0] = m
	}
	side := 0
	if b.colorToMove == Black {
		side = 1
	}
	score := &h.history[side][m.from][m.to]
	*score = min(historyLimit, *score+depth*depth)
}

func (h *moveHeuristics) score(b Board, m Move, ply int) int {
	if tactical := moveOrderScore(b, m); tactical > 0 {
		return 3*historyLimit + tactical
	}
	if ply >= 0 && ply < len(h.killers) {
		if m == h.killers[ply][0] {
			return 2*historyLimit + 2
		}
		if m == h.killers[ply][1] {
			return 2*historyLimit + 1
		}
	}
	side := 0
	if b.colorToMove == Black {
		side = 1
	}
	return h.history[side][m.from][m.to]
}

func (h *moveHeuristics) order(b Board, moves []Move, ply int) {
	sort.SliceStable(moves, func(i, j int) bool {
		return h.score(b, moves[i], ply) > h.score(b, moves[j], ply)
	})
}

// promotions first then most valuable victim and least valuable attacker
// keep the original order for ties
func orderMoves(b Board, moves []Move) {
	sort.SliceStable(moves, func(i, j int) bool {
		return moveOrderScore(b, moves[i]) > moveOrderScore(b, moves[j])
	})
}

func moveOrderScore(b Board, m Move) int {

	ranks := [7]int{0, 6, 1, 2, 3, 4, 5}
	victim := b.squares[m.to].pType()
	if m.isEnPassant {
		victim = Pawn
	}
	score := 0
	if victim != None {
		score = 10*ranks[victim] - ranks[b.squares[m.from].pType()]
	}
	if m.promotion != None {
		score += 100 + ranks[m.promotion]
	}
	return score
}
