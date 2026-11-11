package main

// counts leaf nodes and ignores draw rules
// zero depth counts as one node and negative depth panics
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
