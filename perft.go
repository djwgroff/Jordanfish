package main

// Perft counts legal move sequences of exactly depth plies without changing b.
// Depth zero counts the current position as one node. Like standard perft, it
// ignores repetition, move-count draws, and insufficient material.
// A negative depth panics.
func Perft(b Board, depth int) uint64 {
	if depth < 0 {
		panic("perft: negative depth")
	}
	if depth == 0 {
		return 1
	}

	var nodes uint64
	for _, m := range GenerateLegalMoves(b) {
		next := b
		next.makeMove(m)
		nodes += Perft(next, depth-1)
	}
	return nodes
}
