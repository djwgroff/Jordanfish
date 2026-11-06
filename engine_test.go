package main

import "testing"

// An unpruned reference verifies that pruning preserves the minimax result.
func referenceScore(b Board, depth, ply int) int {
	moves := GenerateLegalMoves(b)
	if len(moves) == 0 {
		if !kingAttacked(b, b.colorToMove) {
			return 0
		}
		if b.colorToMove == White {
			return -mateScore + ply
		}
		return mateScore - ply
	}
	if b.insufficientMaterial() || b.halfmove >= 150 {
		return 0
	}
	if depth == 0 {
		return evaluate(b)
	}
	best := infinity
	if b.colorToMove == White {
		best = -infinity
	}
	for _, m := range moves {
		next := b
		next.makeMove(m)
		score := referenceScore(next, depth-1, ply+1)
		if b.colorToMove == White {
			best = max(best, score)
		} else {
			best = min(best, score)
		}
	}
	return best
}

func TestEngineSearch(t *testing.T) {
	for _, fen := range []string{
		startFEN,
		"7k/8/8/3q4/8/8/8/K2R4 w - - 0 1",
		"k2r4/8/8/8/3Q4/8/8/7K b - - 0 1",
		"7k/8/5KQ1/8/8/8/8/8 w - - 0 1",
		"8/8/8/8/8/5kq1/8/7K b - - 0 1",
		"7k/P7/8/8/8/8/8/7K w - - 0 1",
	} {
		b, err := boardFromFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		original := b
		got, m, ok := minimax(b, 3, -infinity, infinity, 0)
		want := referenceScore(b, 3, 0)
		if got != want {
			t.Fatalf("%s: got %d, want %d", fen, got, want)
		}
		if !ok {
			t.Fatal("no best move")
		}
		legal := false
		for _, candidate := range GenerateLegalMoves(b) {
			if candidate == m {
				legal = true
			}
		}
		if !legal {
			t.Fatal("illegal engine move")
		}
		next := b
		next.makeMove(m)
		if referenceScore(next, 2, 1) != want {
			t.Fatal("selected move does not achieve best score")
		}
		if b != original {
			t.Fatal("search changed board")
		}
	}
}

func TestEngineTactics(t *testing.T) {
	for _, tc := range []struct{ fen, want string }{
		{"7k/8/8/3q4/8/8/8/K2R4 w - - 0 1", "d1d5"},
		{"k2r4/8/8/8/3Q4/8/8/7K b - - 0 1", "d8d4"},
		{"7k/P7/8/8/8/8/8/7K w - - 0 1", "a7a8q"},
	} {
		b, err := boardFromFEN(tc.fen)
		if err != nil {
			t.Fatal(err)
		}
		m, ok := BestMove(b, 2)
		if !ok || m.String() != tc.want {
			t.Fatalf("got %v, want %s", m, tc.want)
		}
	}
	for _, fen := range []string{
		"7k/8/5KQ1/8/8/8/8/8 w - - 0 1",
		"8/8/8/8/8/5kq1/8/7K b - - 0 1",
	} {
		b, err := boardFromFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		m, ok := BestMove(b, 1)
		if !ok {
			t.Fatal("no mating move")
		}
		b.makeMove(m)
		if len(GenerateLegalMoves(b)) != 0 || !kingAttacked(b, b.colorToMove) {
			t.Fatal("missed mate in one at horizon")
		}
	}
}

func TestEngineTerminalPositions(t *testing.T) {
	for _, fen := range []string{
		"7k/6Q1/5K2/8/8/8/8/8 b - - 0 1",
		"7k/5K2/6Q1/8/8/8/8/8 b - - 0 1",
		"7k/8/8/8/8/8/8/7K w - - 0 1",
		"7k/8/8/8/8/8/8/R6K w - - 150 1",
	} {
		b, err := boardFromFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := BestMove(b, 2); ok {
			t.Fatal("move returned in finished position")
		}
	}
	b, _ := boardFromFEN(startFEN)
	for _, depth := range []int{0, -1} {
		if _, ok := BestMove(b, depth); ok {
			t.Fatal("invalid depth accepted")
		}
	}
}

func TestEngineLooksAhead(t *testing.T) {
	b, err := boardFromFEN("7k/8/3p4/4r3/8/8/8/K3Q3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	shallow, _ := BestMove(b, 1)
	deeper, _ := BestMove(b, 2)
	if shallow.String() != "e1e5" || deeper.String() == "e1e5" {
		t.Fatal("deeper search should avoid trading the queen for a protected rook")
	}
}
