# Jordanfish

A Go chess engine with UCI support.

- Minimax with alpha-beta pruning, iterative deepening, and quiescence search
- Transposition table, killer moves, and history ordering
- Piece-square tables, endgame king evaluation, and pawn structure evaluation
- UCI reports for depth, score, nodes, time, NPS, and principal variation

Build and test:

```sh
go build .
go test ./...
```

Select the executable in Cute Chess using UCI. Search defaults to depth 6.
Time management and interruptible search are not supported.

For terminal play, run `go run . -cli`. Enter moves like `e2e4` or `a7a8q`.
Use `engine [depth]` for an engine move, `draw` to offer or accept a draw,
`claim [move]` to claim a draw, and `resign` or `quit` to finish.
