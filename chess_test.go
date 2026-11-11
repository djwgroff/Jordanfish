package main

import (
	"bytes"
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

// unpruned minimax used to check alpha beta results
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
		return quiescence(b, -infinity, infinity, ply, 0)
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
	for _, tc := range []struct {
		fen, want string
		depth     int
	}{
		{"7k/8/8/3q4/8/8/8/K2R4 w - - 0 1", "d1d5", 2},
		{"k2r4/8/8/8/3Q4/8/8/7K b - - 0 1", "d8d4", 2},
		{"7k/P7/8/8/8/8/8/7K w - - 0 1", "a7a8q", 1},
	} {
		b, err := boardFromFEN(tc.fen)
		if err != nil {
			t.Fatal(err)
		}
		m, ok := BestMove(b, tc.depth)
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
	if shallow.String() == "e1e5" || deeper.String() == "e1e5" {
		t.Fatal("quiescence should catch the recapture even at depth one")
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
		{King, "a1", "d4"},
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

func uciTranscript(t *testing.T, input string) string {
	t.Helper()
	var output bytes.Buffer
	if err := runUCI(strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func TestUCIHandshake(t *testing.T) {
	got := uciTranscript(t, "\nuci\nisready\ndebug on\nstop\nisready\nquit\nuci\n")
	want := "id name Jordanfish\noption name Depth type spin default 6 min 1 max 64\nuciok\nreadyok\nreadyok\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got := uciTranscript(t, ""); got != "" {
		t.Fatal("output before handshake")
	}
}

func TestUCIPosition(t *testing.T) {
	for _, tc := range []struct{ command, fen string }{
		{"startpos moves e2e4 c7c5 g1f3", "rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2"},
		{"fen r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1 moves e1g1 e8c8", "2kr3r/8/8/8/8/8/8/R4RK1 w - - 2 2"},
		{"fen 7k/8/8/3pP3/8/8/8/7K w - d6 0 2 moves e5d6", "7k/8/3P4/8/8/8/8/7K b - - 0 2"},
		{"fen 7k/P7/8/8/8/8/8/7K w - - 0 1 moves a7a8n", "N6k/8/8/8/8/8/8/7K b - - 0 1"},
	} {
		got, err := parseUCIPosition(strings.Fields(tc.command))
		if err != nil {
			t.Fatal(err)
		}
		want, err := boardFromFEN(tc.fen)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("incorrect position for %s", tc.command)
		}
	}
	for _, command := range []string{"", "bogus", "fen", "fen 8/8/8/8/8/8/8/8 w - - 0 1", "startpos garbage", "startpos moves e7e5", "startpos moves e2e4 e7e5 e2"} {
		if _, err := parseUCIPosition(strings.Fields(command)); err == nil {
			t.Fatalf("accepted invalid position %q", command)
		}
	}
}

func TestUCISearch(t *testing.T) {
	input := "setoption name Depth value 1\nposition fen 7k/8/8/3q4/8/8/8/K2R4 w - - 0 1\ngo\ngo depth 2\nposition startpos moves e2e4 e7e5 nonsense\ngo\nquit\n"
	output := uciTranscript(t, input)
	if strings.Count(output, "bestmove d1d5\n") != 3 {
		t.Fatal(output)
	}
	if strings.Count(output, "info depth 1\n") != 2 || strings.Count(output, "info depth 2\n") != 1 {
		t.Fatal(output)
	}
	if !strings.Contains(output, "info string invalid position move nonsense\n") {
		t.Fatal(output)
	}
	for _, fen := range []string{
		"7k/6Q1/5K2/8/8/8/8/8 b - - 0 1",
		"7k/5K2/6Q1/8/8/8/8/8 b - - 0 1",
	} {
		output = uciTranscript(t, "position fen "+fen+"\ngo depth 1\nquit\n")
		if !strings.HasSuffix(output, "bestmove 0000\n") {
			t.Fatal(output)
		}
	}
}

func TestUCINewGame(t *testing.T) {
	output := uciTranscript(t, "setoption name depth value 1\nposition startpos moves e2e4\nucinewgame\nisready\ngo wtime 100 btime 100\nquit\n")
	if !strings.HasPrefix(output, "readyok\ninfo depth 1\n") {
		t.Fatal(output)
	}
	b, _ := boardFromFEN(startFEN)
	words := strings.Fields(output)
	if len(words) != 6 || words[4] != "bestmove" {
		t.Fatal(output)
	}
	if _, err := findMove(words[5], GenerateLegalMoves(b)); err != nil {
		t.Fatalf("new game did not reset board: %s", output)
	}
}

func TestUCIInvalidDepth(t *testing.T) {
	input := "setoption name Depth value 1\nsetoption name Depth value 0\nsetoption name Depth value nope\nsetoption name Depth\ngo depth\ngo depth -1\ngo depth 65\ngo depth nope\nisready\ngo\nquit\n"
	output := uciTranscript(t, input)
	if strings.Count(output, "info string ") != 7 || strings.Count(output, "bestmove ") != 1 || !strings.Contains(output, "readyok\ninfo depth 1\n") {
		t.Fatal(output)
	}
}

func referenceQuiescence(b Board, ply, extraPly int) int {
	moves := GenerateLegalMoves(b)
	if score, done := terminalScore(b, moves, ply); done {
		return score
	}
	best := evaluate(b)
	if extraPly >= maxQuiescencePly {
		return best
	}
	inCheck := kingAttacked(b, b.colorToMove)
	if inCheck {
		best = infinity
		if b.colorToMove == White {
			best = -infinity
		}
	}
	for _, m := range moves {
		if !inCheck && b.squares[m.to] == None && !m.isEnPassant && m.promotion == None {
			continue
		}
		next := b
		next.makeMove(m)
		score := referenceQuiescence(next, ply+1, extraPly+1)
		if b.colorToMove == White {
			best = max(best, score)
		} else {
			best = min(best, score)
		}
	}
	return best
}

func TestQuiescence(t *testing.T) {
	for _, fen := range []string{
		startFEN,
		"7k/8/8/3q4/8/8/8/K2R4 w - - 0 1",
		"k2r4/8/8/8/3Q4/8/8/7K b - - 0 1",
		"7k/8/8/3pP3/8/8/8/7K w - d6 0 1",
		"7k/P7/8/8/8/8/8/7K w - - 0 1",
		"1k4r1/8/8/8/8/8/8/R5K1 w - - 0 1",
	} {
		b, err := boardFromFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		original := b
		want := referenceQuiescence(b, 0, 0)
		got := quiescence(b, -infinity, infinity, 0, 0)
		if got != want {
			t.Fatalf("%s: got %d want %d", fen, got, want)
		}
		if b != original {
			t.Fatal("quiescence changed board")
		}
	}
}

func TestQuiescenceTactics(t *testing.T) {
	for _, tc := range []struct {
		fen   string
		color Piece
	}{
		{"7k/8/8/3q4/8/8/8/K2R4 w - - 0 1", White},
		{"k2r4/8/8/8/3Q4/8/8/7K b - - 0 1", Black},
		{"7k/8/8/3pP3/8/8/8/7K w - d6 0 1", White},
		{"7k/P7/8/8/8/8/8/7K w - - 0 1", White},
	} {
		b, err := boardFromFEN(tc.fen)
		if err != nil {
			t.Fatal(err)
		}
		static := evaluate(b)
		got := quiescence(b, -infinity, infinity, 0, 0)
		if tc.color == White && got <= static || tc.color == Black && got >= static {
			t.Fatalf("missed tactical gain: %s", tc.fen)
		}
	}
}

func TestQuiescenceCheckEvasions(t *testing.T) {
	b, err := boardFromFEN("1k4r1/8/8/8/8/8/5P1P/R4RK1 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range GenerateLegalMoves(b) {
		if b.squares[m.to] != None || m.isEnPassant || m.promotion != None {
			t.Fatal("test needs quiet evasions only")
		}
	}
	next := b
	move, err := findMove("g1h1", GenerateLegalMoves(b))
	if err != nil {
		t.Fatal(err)
	}
	next.makeMove(move)
	want := evaluate(next)
	if want >= evaluate(b) {
		t.Fatal("test needs a forced positional loss")
	}
	for _, window := range [][2]int{{-infinity, infinity}, {-infinity, -5}} {
		if got := quiescence(b, window[0], window[1], 0, 0); got != want {
			t.Fatalf("check evasion got %d want %d", got, want)
		}
	}
}

func TestQuiescenceTerminalAndLimit(t *testing.T) {
	for _, tc := range []struct {
		fen  string
		want int
	}{
		{"7k/6Q1/5K2/8/8/8/8/8 b - - 0 1", mateScore - 7},
		{"7k/5K2/6Q1/8/8/8/8/8 b - - 0 1", 0},
		{"7k/8/8/8/8/8/8/R6K w - - 150 1", 0},
		{"7k/8/8/8/8/8/8/7K w - - 0 1", 0},
	} {
		b, err := boardFromFEN(tc.fen)
		if err != nil {
			t.Fatal(err)
		}
		for _, extra := range []int{0, maxQuiescencePly} {
			if got := quiescence(b, -infinity, infinity, 7, extra); got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		}
	}
	b, _ := boardFromFEN("7k/8/8/3q4/8/8/8/K2R4 w - - 0 1")
	if quiescence(b, -infinity, infinity, 0, maxQuiescencePly) != evaluate(b) {
		t.Fatal("extension cap not respected")
	}
}

func TestTranspositionScores(t *testing.T) {
	for _, fen := range []string{
		startFEN,
		"7k/8/8/3q4/8/8/8/K2R4 w - - 0 1",
		"k2r4/8/8/8/3Q4/8/8/7K b - - 0 1",
		"7k/8/5KQ1/8/8/8/8/8 w - - 0 1",
	} {
		b, err := boardFromFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		original := b
		want, _, wantOK := search(b, 3, -infinity, infinity, 0, nil)
		table := make(transpositionTable)
		search(b, 1, -infinity, infinity, 0, table)
		got, move, ok := search(b, 3, -infinity, infinity, 0, table)
		if got != want || ok != wantOK {
			t.Fatalf("cached search got %d want %d for %s", got, want, fen)
		}
		if entry, exists := table[tableKey(b)]; !exists || entry.depth != 3 || entry.bound != exactScore {
			t.Fatal("root result not cached")
		}
		again, cachedMove, _ := search(b, 3, -infinity, infinity, 0, table)
		if again != got || cachedMove != move {
			t.Fatal("cached result changed")
		}
		next := b
		next.makeMove(move)
		if referenceScore(next, 2, 1) != want {
			t.Fatal("cached move does not achieve best score")
		}
		if b != original {
			t.Fatal("search changed board")
		}
	}
}

func TestTranspositionBounds(t *testing.T) {
	b, err := boardFromFEN(startFEN)
	if err != nil {
		t.Fatal(err)
	}
	want, _, _ := search(b, 3, -infinity, infinity, 0, nil)
	for _, tc := range []struct {
		alpha, beta int
		bound       scoreBound
	}{
		{want - 2, want - 1, lowerBound},
		{want + 1, want + 2, upperBound},
	} {
		table := make(transpositionTable)
		score, _, _ := search(b, 3, tc.alpha, tc.beta, 0, table)
		if table[tableKey(b)].bound != tc.bound {
			t.Fatal("incorrect bound flag")
		}
		if tc.bound == lowerBound && score < tc.beta || tc.bound == upperBound && score > tc.alpha {
			t.Fatal("incorrect cutoff bound")
		}
		got, _, _ := search(b, 3, -infinity, infinity, 0, table)
		if got != want {
			t.Fatalf("bound reused as exact score: got %d want %d", got, want)
		}
	}
}

func TestTranspositionMateDistance(t *testing.T) {
	for _, fen := range []string{
		"7k/8/5KQ1/8/8/8/8/8 w - - 0 1",
		"8/8/8/8/8/5kq1/8/7K b - - 0 1",
	} {
		b, err := boardFromFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		table := make(transpositionTable)
		score, _, _ := search(b, 2, -infinity, infinity, 0, table)
		want := mateScore - 1
		if b.colorToMove == Black {
			want = -want
		}
		if score != want {
			t.Fatal("missed mate in one")
		}
		cached, _, _ := search(b, 2, -infinity, infinity, 5, table)
		fresh, _, _ := search(b, 2, -infinity, infinity, 5, nil)
		if cached != fresh {
			t.Fatal("cached mate distance depends on old root")
		}
	}
}

func TestTranspositionState(t *testing.T) {
	b, err := boardFromFEN(startFEN)
	if err != nil {
		t.Fatal(err)
	}
	other := b
	other.fullmove++
	if tableKey(other) != tableKey(b) {
		t.Fatal("fullmove counter prevents transpositions")
	}
	for _, change := range []func(*Board){
		func(b *Board) { b.colorToMove = Black },
		func(b *Board) { b.castling = 0 },
		func(b *Board) { b.enPassant = 20 },
		func(b *Board) { b.halfmove++ },
		func(b *Board) { b.squares[48] = None },
	} {
		other = b
		change(&other)
		if tableKey(other) == tableKey(b) {
			t.Fatal("distinct search states share a key")
		}
	}
	b, err = boardFromFEN("7k/8/8/8/8/8/8/R6K w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	table := make(transpositionTable)
	search(b, 1, -infinity, infinity, 0, table)
	b.halfmove = 150
	score, _, ok := search(b, 1, -infinity, infinity, 0, table)
	if score != 0 || ok {
		t.Fatal("cache hid the move count draw")
	}
}

func TestTranspositionCapacity(t *testing.T) {
	table := make(transpositionTable)
	for i := 0; i < tableCapacity+1; i++ {
		table.store(Board{halfmove: i}, tableEntry{depth: 2})
	}
	if len(table) != tableCapacity {
		t.Fatal("table exceeded capacity")
	}
	key := Board{}
	table.store(key, tableEntry{depth: 3, score: 10})
	table.store(key, tableEntry{depth: 1, score: 20})
	if table[key].score != 10 {
		t.Fatal("shallower entry replaced deeper result")
	}
}

func TestIterativeDeepening(t *testing.T) {
	for _, fen := range []string{
		startFEN,
		"7k/8/8/3q4/8/8/8/K2R4 w - - 0 1",
		"k2r4/8/8/8/3Q4/8/8/7K b - - 0 1",
		"7k/8/3p4/4r3/8/8/8/K3Q3 w - - 0 1",
		"7k/8/5KQ1/8/8/8/8/8 w - - 0 1",
		"8/8/8/8/8/5kq1/8/7K b - - 0 1",
	} {
		b, err := boardFromFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		original := b
		for depth := 1; depth <= 3; depth++ {
			move, ok := BestMove(b, depth)
			if !ok {
				t.Fatal("iterative search returned no move")
			}
			legal := false
			for _, candidate := range GenerateLegalMoves(b) {
				if candidate == move {
					legal = true
				}
			}
			if !legal {
				t.Fatal("iterative search returned an illegal move")
			}
			want := referenceScore(b, depth, 0)
			next := b
			next.makeMove(move)
			if got := referenceScore(next, depth-1, 1); got != want {
				t.Fatalf("depth %d move %s scores %d want %d for %s", depth, move, got, want, fen)
			}
			if b != original {
				t.Fatal("iterative search changed board")
			}
		}
	}
}

func BenchmarkGenerateLegalMoves(b *testing.B) {
	for _, tc := range []struct{ name, fen string }{
		{"start", startFEN},
		{"kiwipete", "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1"},
		{"endgame", "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - - 0 1"},
	} {
		b.Run(tc.name, func(b *testing.B) {
			board, err := boardFromFEN(tc.fen)
			if err != nil {
				b.Fatal(err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				GenerateLegalMoves(board)
			}
		})
	}
}

func BenchmarkPerft(b *testing.B) {
	board, err := boardFromFEN(startFEN)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if Perft(board, 3) != 8902 {
			b.Fatal("incorrect node count")
		}
	}
}

func referenceSquareAttacked(b Board, square int, by Piece) bool {
	r, f := square/8, square%8
	for from, p := range b.squares {
		if p == None || p.pColor() != by {
			continue
		}
		dr, df := r-from/8, f-from%8
		switch p.pType() {
		case Pawn:
			step := -1
			if by == Black {
				step = 1
			}
			if dr == step && abs(df) == 1 {
				return true
			}
		case Knight:
			if abs(dr)*abs(df) == 2 {
				return true
			}
		case King:
			if abs(dr) <= 1 && abs(df) <= 1 {
				return true
			}
		default:
			diagonal := abs(dr) == abs(df) && dr != 0
			straight := (dr == 0) != (df == 0)
			if !((diagonal && (p.pType() == Bishop || p.pType() == Queen)) || (straight && (p.pType() == Rook || p.pType() == Queen))) {
				continue
			}
			rstep, fstep := 0, 0
			if dr != 0 {
				rstep = dr / abs(dr)
			}
			if df != 0 {
				fstep = df / abs(df)
			}
			clear := true
			for rr, ff := from/8+rstep, from%8+fstep; rr != r || ff != f; rr, ff = rr+rstep, ff+fstep {
				if b.squares[rr*8+ff] != None {
					clear = false
					break
				}
			}
			if clear {
				return true
			}
		}
	}
	return false
}

func TestMoveGenerationMatchesReference(t *testing.T) {
	seed := uint32(1)
	for _, fen := range []string{
		startFEN,
		"r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1",
		"7k/8/8/3pP3/8/8/8/7K w - d6 0 1",
		"7k/P7/8/8/8/8/8/7K w - - 0 1",
	} {
		board, err := boardFromFEN(fen)
		if err != nil {
			t.Fatal(err)
		}
		for ply := 0; ply < 64; ply++ {
			original := board
			for square := 0; square < 64; square++ {
				for _, by := range []Piece{White, Black} {
					if board.squares[square] == by|King {
						continue
					}
					if squareAttacked(board, square, by) != referenceSquareAttacked(board, square, by) {
						t.Fatalf("attack mismatch at square %d by color %d at ply %d from %s", square, by, ply, fen)
					}
				}
			}
			var want []Move
			for _, m := range GenerateMoves(board) {
				next := board
				next.makeMove(m)
				for king, p := range next.squares {
					if p == board.colorToMove|King {
						if !referenceSquareAttacked(next, king, opposite(board.colorToMove)) {
							want = append(want, m)
						}
						break
					}
				}
			}
			got := GenerateLegalMoves(board)
			if len(got) != len(want) {
				t.Fatalf("got %d moves want %d at ply %d from %s", len(got), len(want), ply, fen)
			}
			for i := range got {
				if got[i] != want[i] {
					t.Fatal("legal moves or order changed")
				}
			}
			if board != original {
				t.Fatal("move generation changed board")
			}
			if len(got) == 0 {
				break
			}
			seed = seed*1664525 + 1013904223
			board.makeMove(got[int(seed)%len(got)])
		}
	}
}

func TestKingPhaseBlend(t *testing.T) {
	for _, queens := range []int{0, 3, 6, 7} {
		var before Board
		before.squares[60], before.squares[4] = White|King, Black|King
		for i := 0; i < queens; i++ {
			before.squares[8+i] = White | Queen
		}
		after := before
		after.squares[60], after.squares[27] = None, White|King
		phase := min(queens*4, 24)
		startBonus := (pieceSquareTables[King][60]*phase + endgameKingTable[60]*(24-phase)) / 24
		endBonus := (pieceSquareTables[King][27]*phase + endgameKingTable[27]*(24-phase)) / 24
		if got := evaluate(after) - evaluate(before); got != endBonus-startBonus {
			t.Fatalf("phase %d: got %d want %d", phase, got, endBonus-startBonus)
		}
		if phase == 0 && evaluate(after) <= evaluate(before) {
			t.Fatal("endgame king should favor the center")
		}
		if phase == 24 && evaluate(after) >= evaluate(before) {
			t.Fatal("opening king should favor shelter")
		}
		var mirroredBefore, mirroredAfter Board
		for square, p := range before.squares {
			if p != None {
				mirroredBefore.squares[square^56] = opposite(p.pColor()) | p.pType()
			}
		}
		for square, p := range after.squares {
			if p != None {
				mirroredAfter.squares[square^56] = opposite(p.pColor()) | p.pType()
			}
		}
		if evaluate(mirroredBefore) != -evaluate(before) || evaluate(mirroredAfter) != -evaluate(after) {
			t.Fatal("king blend is not symmetric")
		}
	}
}

func TestBishopPairBonus(t *testing.T) {
	for _, color := range []Piece{White, Black} {
		var b Board
		sign := 1
		if color == Black {
			sign = -1
		}
		materialAndSquares := 0
		for i, square := range []int{58, 61, 42} {
			target := square
			if color == Black {
				target ^= 56
			}
			b.squares[target] = color | Bishop
			materialAndSquares += 330 + pieceSquareTables[Bishop][square]
			want := materialAndSquares
			if i >= 1 {
				want += 30
			}
			if got := evaluate(b); got != sign*want {
				t.Fatalf("%d bishops of color %d: got %d want %d", i+1, color, got, sign*want)
			}
		}
	}
	var b Board
	b.squares[58], b.squares[61] = White|Bishop, White|Bishop
	b.squares[2], b.squares[5] = Black|Bishop, Black|Bishop
	if evaluate(b) != 0 {
		t.Fatal("equal bishop pairs should balance")
	}
}

func TestPawnPenalties(t *testing.T) {
	for _, tc := range []struct {
		name    string
		squares []string
		penalty int
	}{
		{"no pawns", nil, 0},
		{"isolated", []string{"d4"}, 15},
		{"connected", []string{"d4", "e2"}, 0},
		{"doubled isolated", []string{"d4", "d3"}, 40},
		{"tripled isolated", []string{"d4", "d3", "d2"}, 65},
		{"doubled supported", []string{"d4", "d3", "e2"}, 10},
		{"tripled supported", []string{"d4", "d3", "d2", "e2"}, 20},
		{"edge isolated", []string{"a4"}, 15},
		{"edge supported", []string{"a4", "b2"}, 0},
		{"edges do not connect", []string{"a4", "h4"}, 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, color := range []Piece{White, Black} {
				var b Board
				raw := 0
				for _, s := range tc.squares {
					square, err := coordinateToIndex(s)
					if err != nil {
						t.Fatal(err)
					}
					raw += 100 + pieceSquareTables[Pawn][square]
					if color == Black {
						square ^= 56
					}
					b.squares[square] = color | Pawn
				}
				original := b
				want := raw - tc.penalty
				if color == Black {
					want = -want
				}
				if got := evaluate(b); got != want {
					t.Fatalf("color %d: got %d want %d", color, got, want)
				}
				if b != original {
					t.Fatal("evaluation changed board")
				}
			}
		})
	}
	var b Board
	raw := 0
	for _, tc := range []struct {
		square string
		color  Piece
	}{{"d4", White}, {"e5", Black}, {"f5", Black}} {
		square, _ := coordinateToIndex(tc.square)
		b.squares[square] = tc.color | Pawn
		if tc.color == White {
			raw += 100 + pieceSquareTables[Pawn][square]
		} else {
			raw -= 100 + pieceSquareTables[Pawn][square^56]
		}
	}
	if got := evaluate(b); got != raw-15 {
		t.Fatal("enemy pawns should not prevent isolation")
	}
	start, _ := boardFromFEN(startFEN)
	if evaluate(start) != 0 {
		t.Fatal("starting position should remain balanced")
	}
}

func TestKillerAndHistoryUpdates(t *testing.T) {
	b, err := boardFromFEN(startFEN)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := findMove("g1f3", GenerateLegalMoves(b))
	c, _ := findMove("b1c3", GenerateLegalMoves(b))
	h := new(moveHeuristics)
	h.recordCutoff(b, a, 3, 4)
	h.recordCutoff(b, c, 2, 4)
	h.recordCutoff(b, c, 2, 4)
	if h.killers[4] != [2]Move{c, a} || h.killers[3] != [2]Move{} {
		t.Fatal("killers must be distinct and specific to a ply")
	}
	if h.history[0][a.from][a.to] != 9 || h.history[0][c.from][c.to] != 8 || h.history[1][a.from][a.to] != 0 {
		t.Fatal("history rewards must use depth squared and stay separate by side")
	}
	h.recordCutoff(b, a, 1, 4)
	if h.killers[4] != [2]Move{a, c} {
		t.Fatal("recent killer was not promoted")
	}
	b.colorToMove = Black
	h.recordCutoff(b, Move{from: 6, to: 21}, 2, 5)
	if h.history[1][6][21] != 4 || h.history[0][6][21] != 0 {
		t.Fatal("black history used the wrong side")
	}
	b.colorToMove = White
	h.history[0][a.from][a.to] = historyLimit - 1
	h.recordCutoff(b, a, 3, 128)
	if h.history[0][a.from][a.to] != historyLimit {
		t.Fatal("history should saturate without overflowing")
	}
	h.recordCutoff(b, a, 3, 128)
	if h.history[0][a.from][a.to] != historyLimit {
		t.Fatal("history exceeded its limit")
	}
}

func TestQuietMoveOrdering(t *testing.T) {
	b, err := boardFromFEN("7k/8/8/1p1q4/4r3/2N5/8/K2R4 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	legal := GenerateLegalMoves(b)
	get := func(s string) Move {
		m, err := findMove(s, legal)
		if err != nil {
			t.Fatal(err)
		}
		return m
	}
	h := new(moveHeuristics)
	h.recordCutoff(b, get("a1b1"), 1, 3)
	h.recordCutoff(b, get("d1d2"), 1, 3)
	historyMove := get("d1d3")
	h.history[0][historyMove.from][historyMove.to] = historyLimit
	moves := []Move{get("c3a4"), historyMove, get("a1b1"), get("c3d5"), get("d1d2")}
	h.order(b, moves, 3)
	for i, want := range []string{"c3d5", "d1d2", "a1b1", "d1d3", "c3a4"} {
		if moves[i].String() != want {
			t.Fatalf("got %v", moves)
		}
	}
	h.order(b, moves, 2)
	if moves[1] != historyMove {
		t.Fatal("killers leaked to another ply")
	}
}

func TestTacticalCutoffsDoNotUpdateHistory(t *testing.T) {
	for _, tc := range []struct{ fen, move string }{
		{"7k/8/8/3q4/8/8/8/K2R4 w - - 0 1", "d1d5"},
		{"7k/8/8/3pP3/8/8/8/7K w - d6 0 1", "e5d6"},
		{"7k/P7/8/8/8/8/8/7K w - - 0 1", "a7a8q"},
	} {
		b, err := boardFromFEN(tc.fen)
		if err != nil {
			t.Fatal(err)
		}
		m, err := findMove(tc.move, GenerateLegalMoves(b))
		if err != nil {
			t.Fatal(err)
		}
		h := new(moveHeuristics)
		h.recordCutoff(b, m, 3, 2)
		if *h != (moveHeuristics{}) {
			t.Fatal("tactical cutoff changed quiet move heuristics")
		}
	}
}

func TestSearchRecordsQuietCutoffs(t *testing.T) {
	for _, color := range []Piece{White, Black} {
		b, err := boardFromFEN(startFEN)
		if err != nil {
			t.Fatal(err)
		}
		b.colorToMove = color
		alpha, beta, side := -infinity, -infinity+1, 0
		if color == Black {
			alpha, beta, side = infinity-1, infinity, 1
		}
		h := new(moveHeuristics)
		_, move, ok := searchWithOrdering(b, 1, alpha, beta, 0, nil, h)
		if !ok || h.killers[0][0] != move || h.history[side][move.from][move.to] != 1 {
			t.Fatal("search did not record a quiet cutoff")
		}
	}
}
