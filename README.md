# Jordanfish

A chess engine.
Uses a recursive minimax search, alpha-beta pruning, and basic move ordering. 

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
