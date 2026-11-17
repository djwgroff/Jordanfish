package main

type moveBuffers struct {
	plies [][]Move
}

func (buffers *moveBuffers) generate(b Board, ply int, tacticalOnly bool) []Move {
	for len(buffers.plies) <= ply {
		buffers.plies = append(buffers.plies, nil)
	}
	moves := generateLegalMoves(b, tacticalOnly, buffers.plies[ply])
	buffers.plies[ply] = moves
	return moves
}

func (h *moveHeuristics) ensureBuffers() *moveBuffers {
	if h.buffers == nil {
		h.buffers = new(moveBuffers)
	}
	return h.buffers
}
