package main

import (
	"fmt"
	"io"
	"time"
)

// scores favor white and nodes include quiescence and transposition hits
type SearchInfo struct {
	Depth   int
	Score   int
	Nodes   uint64
	Elapsed time.Duration
	PV      []Move
}

func (info SearchInfo) NPS() uint64 {
	if info.Elapsed <= 0 {
		return 0
	}
	return uint64(float64(info.Nodes) / info.Elapsed.Seconds())
}

func principalVariation(b Board, first Move, depth int, table transpositionTable) []Move {
	var pv []Move
	move := first
	for len(pv) < depth {
		legal := false
		for _, candidate := range GenerateLegalMoves(b) {
			if candidate == move {
				legal = true
				break
			}
		}
		if !legal {
			break
		}
		pv = append(pv, move)
		b.makeMove(move)
		entry, ok := table[tableKey(b)]
		if !ok || entry.bound != exactScore || entry.depth < depth-len(pv) {
			break
		}
		move = entry.move
	}
	return pv
}

func writeSearchInfo(output io.Writer, color Piece, info SearchInfo) {
	score := info.Score
	if color == Black {
		score = -score
	}
	fmt.Fprint(output, "info depth ", info.Depth, " score ")
	if abs(score) >= mateScore-1024 {
		moves := (mateScore - abs(score) + 1) / 2
		if score < 0 {
			moves = -moves
		}
		fmt.Fprint(output, "mate ", moves)
	} else {
		fmt.Fprint(output, "cp ", score)
	}
	fmt.Fprintf(output, " nodes %d time %d nps %d", info.Nodes, info.Elapsed.Milliseconds(), info.NPS())
	if len(info.PV) > 0 {
		fmt.Fprint(output, " pv")
		for _, move := range info.PV {
			fmt.Fprint(output, " ", move)
		}
	}
	fmt.Fprintln(output)
}
