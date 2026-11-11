# Jordanfish

A chess engine.
Uses a recursive minimax search, alpha-beta pruning, basic move ordering, and piece-square tables.

Build with `go build -o jordanfish .`. The executable defaults to UCI mode. In
Cute Chess, add it as an engine using the UCI protocol, with no arguments.
For the terminal game below, run `./jordanfish -cli` or `go run . -cli`.

UCI supports `uci`, `isready`, `ucinewgame`, `position startpos` or `position fen`
with an optional `moves` list, `setoption name Depth value N`, `go depth N`, and
`quit`. Depth defaults to 3 and accepts 1–64. Use fixed-depth games in Cute Chess.
Clock and other search limits are ignored in favor of the configured depth.
Search runs synchronously, so `stop`, `isready`, and `quit` received during a
search are only processed after it finishes. There is no pondering or time management.

Enter moves as `e2e4`. Use `e1g1` / `e1c1` for White's castling and `e8g8` /
`e8c8` for Black's. Promotions require a suffix: `q`, `r`, `b`, or `n`, such as
`a7a8q`. Coordinates stay the same when the board rotates.

Commands:

- `engine [depth]`: have the engine play the current side's move. Depth defaults
  to 3 plies (range 1–6); higher depths take longer. To play White against the
  engine, enter your move, then `engine` for Black's reply.
- `quit`: exit; `resign`: concede the game.
- `draw`: offer a draw, then play your move. Your opponent can accept with `draw`
  or decline by playing a move.
- `claim`: claim an available threefold-repetition or fifty-move draw.
- `claim <move>`: claim a draw that your intended legal move would produce.

Checkmate, stalemate, common insufficient-material positions, fivefold repetition,
and the seventy-five-move rule end the game automatically. Invalid input is
rejected; end-of-input exits cleanly.
