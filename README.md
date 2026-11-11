# Jordanfish

A chess engine.
Uses a recursive minimax search, alpha-beta pruning, basic move ordering, and piece-square tables.
King evaluation blends shelter in the opening with centralization in the endgame
based on remaining material. Each side gets a 30 centipawn bonus for a bishop pair.
Pawn structure costs 10 centipawns per extra pawn on a file and 15 per isolated
pawn with no friendly pawns on either neighboring file. The penalties can stack.
Quiescence search extends captures and promotions at the depth limit and searches
all legal replies when in check, with a safety cap of 16 extra plies.
A transposition table caches search scores and best moves with up to 65536 entries
per search. It keeps exact scores and alpha-beta bounds and tries cached moves first.
Iterative deepening searches each depth up to the requested limit and keeps the
same table between iterations to improve move ordering.

UCI supports `uci`, `isready`, `ucinewgame`, `position startpos` or `position fen`
with an optional `moves` list, `setoption name Depth value N`, `go depth N`, and
`quit`. Depth defaults to 3 and accepts 1–64.

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
