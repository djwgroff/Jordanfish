package main

import "sort"

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
