package main

import "testing"

func TestCaptureOrdering(t *testing.T) {
	for _, tc := range []struct {
		fen         string
		input, want []string
	}{
		{"7k/8/8/1p1q4/4r3/2N5/8/K2R4 w - - 0 1",
			[]string{"a1b1", "c3b5", "c3e4", "d1d5", "c3d5"},
			[]string{"c3d5", "d1d5", "c3e4", "c3b5", "a1b1"}},
		{"k2r4/8/2n5/4R3/1P1Q4/8/8/7K b - - 0 1",
			[]string{"a8b8", "c6b4", "c6e5", "d8d4", "c6d4"},
			[]string{"c6d4", "d8d4", "c6e5", "c6b4", "a8b8"}},
		{"7k/8/8/3pP3/8/8/8/7K w - d6 0 1",
			[]string{"h1g1", "e5d6", "e5e6"},
			[]string{"e5d6", "h1g1", "e5e6"}},
		{"7k/P7/8/8/8/8/8/7K w - - 0 1",
			[]string{"h1g1", "a7a8n", "a7a8b", "a7a8r", "a7a8q"},
			[]string{"a7a8q", "a7a8r", "a7a8b", "a7a8n", "h1g1"}},
		{"r6k/1P6/8/8/8/8/8/7K w - - 0 1",
			[]string{"b7b8q", "b7a8q", "h1g1"},
			[]string{"b7a8q", "b7b8q", "h1g1"}},
	} {
		b, err := boardFromFEN(tc.fen)
		if err != nil {
			t.Fatal(err)
		}
		original := b
		legal := GenerateLegalMoves(b)
		moves := make([]Move, len(tc.input))
		for i, s := range tc.input {
			moves[i], err = findMove(s, legal)
			if err != nil {
				t.Fatalf("%s: %v", s, err)
			}
		}
		orderMoves(b, moves)
		for i, m := range moves {
			if m.String() != tc.want[i] {
				t.Fatalf("got %v, want %v", moves, tc.want)
			}
		}
		if b != original {
			t.Fatal("ordering changed board")
		}
	}
}
