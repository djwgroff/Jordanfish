# Jordanfish

A small, two-player terminal chess game written in Go. Requires Go 1.22 or newer;
no external dependencies.

```sh
go run .
```

In the cloud environment, first run `source /workspace/.jordanfish-tools/activate.sh`
to select the Go compiler.

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

The engine uses material evaluation and minimax with alpha-beta pruning only.
It recognizes checkmate, stalemate, insufficient material, and the seventy-five-move
rule during search; repetition history and optional draw claims are not searched.
