package main

import "sort"

// orderMoves searches promotions and captures first. Captures prefer the most
// valuable victim, then the least valuable attacker (MVV-LVA). Quiet moves and
// other ties retain their generation order.
func orderMoves(b Board, moves []Move) {
	sort.SliceStable(moves, func(i, j int) bool {
		return moveOrderScore(b, moves[i]) > moveOrderScore(b, moves[j])
	})
}

func moveOrderScore(b Board, m Move) int {
	// Relative piece ranks, including the king as the most costly attacker.
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
