package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func position(t *testing.T, fen string) Board {
	t.Helper()
	b, err := boardFromFEN(fen)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func move(t *testing.T, b *Board, s string) {
	t.Helper()
	m, err := findMove(s, GenerateLegalMoves(*b))
	if err != nil {
		t.Fatalf("%s: %v", s, err)
	}
	b.makeMove(m)
}
func hasMove(b Board, s string) bool { _, err := findMove(s, GenerateLegalMoves(b)); return err == nil }
func TestPerft(t *testing.T) {
	for _, tc := range []struct {
		name, fen   string
		depth, want int
	}{
		{"start-depth-zero", startFEN, 0, 1},
		{"start-depth-one", startFEN, 1, 20},
		{"start-depth-two", startFEN, 2, 400},
		{"start-depth-three", startFEN, 3, 8902},
		{"start", startFEN, 4, 197281},
		{"stalemate", "7k/5K2/6Q1/8/8/8/8/8 b - - 0 1", 1, 0},
		{"ignore-draws", "7k/8/8/8/8/8/8/7K w - - 150 1", 1, 3},
		{"kiwipete", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1", 3, 97862},
		{"rook-pawn-endgame", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1", 4, 43238},
		{"promotion-castling", "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1", 3, 9467},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := position(t, tc.fen)
			original := b
			if got := Perft(b, tc.depth); got != uint64(tc.want) {
				t.Fatalf("got %d nodes, want %d", got, tc.want)
			}
			if b != original {
				t.Fatal("Perft changed the board")
			}
		})
	}
}
func TestKingSafety(t *testing.T) {
	b := position(t, "k3r3/8/8/8/8/8/4R3/4K3 w - - 0 1")
	if hasMove(b, "e2f2") {
		t.Fatal("pinned rook allowed to expose king")
	}
	b = position(t, "7k/8/8/3p4/8/4K3/8/8 w - - 0 1")
	if hasMove(b, "e3e4") {
		t.Fatal("king can move onto pawn attack")
	}
	b = position(t, "8/8/8/4k3/8/4K3/8/8 w - - 0 1")
	if hasMove(b, "e3e4") {
		t.Fatal("kings can become adjacent")
	}
}
func TestPawnRules(t *testing.T) {
	b := position(t, startFEN)
	move(t, &b, "e2e4")
	if b.enPassant != 44 || b.halfmove != 0 {
		t.Fatal("double push state incorrect")
	}
	b = position(t, "7k/8/8/8/8/4n3/4P3/7K w - - 0 1")
	if hasMove(b, "e2e4") || hasMove(b, "e2e3") {
		t.Fatal("pawn jumps blocker")
	}
	for _, color := range []string{"w", "b"} {
		fen, from, to, pc := "7k/P7/8/8/8/8/8/7K w - - 0 1", "a7", "a8", White
		if color == "b" {
			fen, from, to, pc = "7k/8/8/8/8/8/p7/7K b - - 0 1", "a2", "a1", Black
		}
		for suffix, typ := range map[string]Piece{"q": Queen, "r": Rook, "b": Bishop, "n": Knight} {
			b = position(t, fen)
			move(t, &b, from+to+suffix)
			sq, _ := coordinateToIndex(to)
			if b.squares[sq] != pc|typ {
				t.Fatal("wrong promotion")
			}
			GenerateLegalMoves(b)
		}
	}
}
func TestEnPassant(t *testing.T) {
	b := position(t, "7k/8/8/3pP3/8/8/8/7K w - d6 0 2")
	previous := b
	move(t, &b, "e5d6")
	if b.squares[27] != None || b.squares[19] != White|Pawn || b.enPassant != -1 {
		t.Fatal("en passant not applied")
	}
	b.unMakeMove(previous)
	if b != previous {
		t.Fatal("undo did not restore en passant state")
	}
	move(t, &b, "h1g1")
	if b.enPassant != -1 {
		t.Fatal("en passant did not expire")
	}
	b = position(t, "7k/8/8/r4pPK/8/8/8/8 w - f6 0 2")
	if hasMove(b, "g5f6") {
		t.Fatal("en passant exposes rook check")
	}
}
func TestCastling(t *testing.T) {
	for _, tc := range []struct {
		fen, move  string
		king, rook int
		color      Piece
	}{
		{"r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", "e1g1", 62, 61, White},
		{"r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1", "e1c1", 58, 59, White},
		{"r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1", "e8g8", 6, 5, Black},
		{"r3k2r/8/8/8/8/8/8/R3K2R b KQkq - 0 1", "e8c8", 2, 3, Black},
	} {
		b := position(t, tc.fen)
		previous := b
		move(t, &b, tc.move)
		if b.squares[tc.king] != tc.color|King || b.squares[tc.rook] != tc.color|Rook {
			t.Fatal("castling pieces misplaced")
		}
		b.unMakeMove(previous)
		if b != previous {
			t.Fatal("castling undo failed")
		}
	}
	for _, fen := range []string{
		"4kr2/8/8/8/8/8/8/4K2R w K - 0 1",  // through check
		"4k1r1/8/8/8/8/8/8/4K2R w K - 0 1", // into check
		"k3r3/8/8/8/8/8/8/4K2R w K - 0 1",  // out of check
		"4k3/8/8/8/8/8/8/4KB1R w K - 0 1",  // blocked
		"4k3/8/8/8/8/8/8/4K2R w - - 0 1",   // no right
	} {
		if hasMove(position(t, fen), "e1g1") {
			t.Fatalf("illegal castle: %s", fen)
		}
	}
	b := position(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	move(t, &b, "h1h8")
	if b.castling&(whiteKingSide|blackKingSide) != 0 {
		t.Fatal("rook move/capture kept castling rights")
	}
	b = position(t, "4k3/8/8/8/8/8/8/4K2R w K - 0 1")
	for _, s := range []string{"h1h2", "e8e7", "h2h1", "e7e8"} {
		move(t, &b, s)
	}
	if hasMove(b, "e1g1") {
		t.Fatal("rook returning home restored castling")
	}
}
func TestUndoCapture(t *testing.T) {
	b := position(t, "7k/8/8/8/8/p7/R7/7K w - - 7 8")
	m, err := findMove("a2a3", GenerateLegalMoves(b))
	if err != nil {
		t.Fatal(err)
	}
	original := b
	previous := b.makeMove(m)
	b.unMakeMove(previous)
	if b != original {
		t.Fatal("capture undo lost state")
	}
}
func TestParseErrors(t *testing.T) {
	for _, s := range []string{"", "e2", "e2e", "i2a3", "a0a1", "a9a8", "e2e4x", "e2e4qq", "é2e4"} {
		if _, err := parseMove(s); err == nil {
			t.Errorf("accepted bad move %q", s)
		}
	}
	for _, s := range []string{"", "8", "9/8/8/8/8/8/8/8", "8/8/8/8/8/8/8/8", "7k/8/8/8/8/8/8/7X", "7k/8/8/8/8/8/8/6KP", "7k/8/8/8/8/8/8/7K x - - 0 1", "7k/8/8/8/8/8/8/7K w K - 0 1", "7k/8/8/8/8/8/8/7K w - e4 0 1", "7k/8/8/8/8/8/8/7K w - - -1 1", "7k/8/8/8/8/8/8/7K w - - 0 0"} {
		if _, err := boardFromFEN(s); err == nil {
			t.Errorf("accepted bad FEN %q", s)
		}
	}
	b := position(t, "7k/8/8/8/8/8/8/7K b - - 42 17")
	if b.colorToMove != Black || b.halfmove != 42 || b.fullmove != 17 {
		t.Fatal("FEN counters not loaded")
	}
}
func TestGameEndings(t *testing.T) {
	b := position(t, startFEN)
	for _, s := range []string{"f2f3", "e7e5", "g2g4", "d8h4"} {
		move(t, &b, s)
	}
	if got := gameResult(b, GenerateLegalMoves(b), 1); got != "Checkmate. Black wins." {
		t.Fatal(got)
	}
	b = position(t, "7k/5K2/6Q1/8/8/8/8/8 b - - 0 1")
	if got := gameResult(b, GenerateLegalMoves(b), 1); got != "Draw by stalemate." {
		t.Fatal(got)
	}
	for _, tc := range []struct {
		fen  string
		want bool
	}{
		{"7k/8/8/8/8/8/8/7K", true},
		{"7k/8/8/8/8/8/8/5B1K", true},
		{"7k/8/8/8/8/8/8/5N1K", true},
		{"7k/8/8/8/8/8/5b2/6BK", true}, // same color bishops
		{"7k/8/8/8/8/8/6b1/6BK", false},
		{"7k/8/8/8/8/8/8/4NN1K", false},
		{"7k/8/8/8/8/8/8/4BN1K", false},
	} {
		if got := position(t, tc.fen).insufficientMaterial(); got != tc.want {
			t.Errorf("material: %s got %v", tc.fen, got)
		}
	}
	b = position(t, startFEN)
	b.halfmove = 100
	if drawClaim(b, 1) == "" || gameResult(b, GenerateLegalMoves(b), 1) != "" {
		t.Fatal("50 moves should be claimable only")
	}
	b.halfmove = 150
	if gameResult(b, GenerateLegalMoves(b), 1) != "Draw by the seventy-five-move rule." {
		t.Fatal("75-move draw missing")
	}
	b.halfmove = 0
	if drawClaim(b, 3) == "" || gameResult(b, GenerateLegalMoves(b), 5) != "Draw by fivefold repetition." {
		t.Fatal("repetition draw missing")
	}
}
func TestRepetitionKey(t *testing.T) {
	b := position(t, startFEN)
	key := b.key()
	for _, s := range []string{"g1f3", "g8f6", "f3g1", "f6g8"} {
		move(t, &b, s)
	}
	if b.key() != key {
		t.Fatal("move counters prevent repetition")
	}
	noEP := position(t, "7k/8/8/3p4/8/8/8/7K w - - 0 1")
	withEP := position(t, "7k/8/8/3p4/8/8/8/7K w - d6 0 1")
	if noEP.key() != withEP.key() {
		t.Fatal("unusable EP target changes repetition")
	}
	withEP = position(t, "7k/8/8/3pP3/8/8/8/7K w - d6 0 1")
	noEP = withEP
	noEP.enPassant = -1
	if noEP.key() == withEP.key() {
		t.Fatal("legal EP capture ignored in repetition")
	}
	withEP = position(t, "7k/8/8/r4pPK/8/8/8/8 w - f6 0 1")
	noEP = withEP
	noEP.enPassant = -1
	if noEP.key() != withEP.key() {
		t.Fatal("pinned EP capture changes repetition")
	}
}
func TestTerminal(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"e2\ni2a3\ne2e4\nquit\n", "Black>"},
		{"", "White>"},
		{"draw\ne2e4\ndraw\n", "Draw by agreement."},
		{"draw\ne2e4\ne7e5\ndraw\ng1f3\ndraw\n", "Draw by agreement."},
		{"resign\n", "White resigns. Black wins."},
		{"f2f3\ne7e5\ng2g4\nd8h4\n", "Checkmate. Black wins."},
		{"g1f3\ng8f6\nf3g1\nf6g8\ng1f3\ng8f6\nf3g1\nclaim f6g8\n", "Draw claimed by threefold repetition."},
		{strings.Repeat("g1f3\ng8f6\nf3g1\nf6g8\n", 4), "Draw by fivefold repetition."},
	} {

		file, err := os.CreateTemp(t.TempDir(), "transcript")
		if err != nil {
			t.Fatal(err)
		}
		original := os.Stdout
		os.Stdout = file
		func() { defer func() { os.Stdout = original }(); err = play(strings.NewReader(tc.input)) }()
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		out, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), tc.want) {
			t.Fatalf("missing %q in transcript", tc.want)
		}
	}
}

// unpruned search used to check alpha beta results
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

func TestEvaluationSymmetry(t *testing.T) {
	for _, fen := range []string{
		startFEN,
		"r3k2r/pp3ppp/2n1bn2/2qp4/3P4/2N1PN2/PPQ1BPPP/R3K2R w KQkq - 0 1",
	} {
		b, err := boardFromFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		original := b
		var mirrored Board
		for square, p := range b.squares {
			if p != None {
				mirrored.squares[square^56] = p.pType() | opposite(p.pColor())
			}
		}
		if evaluate(b) != -evaluate(mirrored) {
			t.Fatal("swapping colors and flipping ranks must negate evaluation")
		}
		if b != original {
			t.Fatal("evaluation changed board")
		}
	}
	b, _ := boardFromFEN(startFEN)
	if evaluate(b) != 0 {
		t.Fatal("starting position should be balanced")
	}
}

func TestPiecePlacement(t *testing.T) {
	for _, tc := range []struct {
		piece         Piece
		worse, better string
	}{
		{Knight, "a1", "d4"},
		{Bishop, "a1", "d4"},
		{Pawn, "e2", "e4"},
		{Rook, "a1", "a7"},
		{Queen, "a1", "d4"},
		{King, "e1", "g1"},
	} {
		worse, _ := coordinateToIndex(tc.worse)
		better, _ := coordinateToIndex(tc.better)
		var before, after Board
		before.squares[worse], after.squares[better] = White|tc.piece, White|tc.piece
		if evaluate(after) <= evaluate(before) {
			t.Errorf("white %d should prefer %s to %s", tc.piece, tc.better, tc.worse)
		}
		before, after = Board{}, Board{}
		before.squares[worse^56], after.squares[better^56] = Black|tc.piece, Black|tc.piece
		if evaluate(after) >= evaluate(before) {
			t.Errorf("black %d improvement must lower White's score", tc.piece)
		}
	}
}

func TestEngineDevelopsKnight(t *testing.T) {
	for _, color := range []Piece{White, Black} {
		b, err := boardFromFEN(startFEN)
		if err != nil {
			t.Fatal(err)
		}
		b.colorToMove = color
		m, ok := BestMove(b, 1)
		if !ok || b.squares[m.from].pType() != Knight {
			t.Fatal("depth-one engine should favor knight development")
		}
		before := evaluate(b)
		b.makeMove(m)
		after := evaluate(b)
		if color == White && after <= before || color == Black && after >= before {
			t.Fatal("engine did not improve piece placement")
		}
	}
}
