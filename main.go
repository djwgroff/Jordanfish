package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func main() {
	cli := flag.Bool("cli", false, "play in the terminal instead of using UCI")
	flag.Parse()
	var err error
	if *cli {
		err = play(os.Stdin)
	} else {
		err = runUCI(os.Stdin, os.Stdout)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func findMove(s string, legal []Move) (Move, error) {
	m, err := parseMove(s)
	if err != nil {
		return Move{}, err
	}
	for _, candidate := range legal {
		if m.from == candidate.from && m.to == candidate.to && m.promotion == candidate.promotion {
			return candidate, nil
		}
	}
	return Move{}, fmt.Errorf("illegal move (promotions need q, r, b, or n, e.g. a7a8q)")
}

func drawClaim(b Board, repetitions int) string {
	if repetitions >= 3 {
		return "Draw claimed by threefold repetition."
	}
	if b.halfmove >= 100 {
		return "Draw claimed by the fifty-move rule."
	}
	return ""
}

func play(input io.Reader) error {
	b, err := boardFromFEN(startFEN)
	if err != nil {
		return err
	}
	repetitions := map[positionKey]int{b.key(): 1}
	drawOffer := None
	scanner := bufio.NewScanner(input)
	fmt.Println("Moves: e2e4, a7a8q. Commands: engine [depth], quit, resign, draw, claim [move].")
	for {
		displayBoard(b, b.colorToMove)
		legal := GenerateLegalMoves(b)
		if result := gameResult(b, legal, repetitions[b.key()]); result != "" {
			fmt.Println(result)
			return nil
		}
		if kingAttacked(b, b.colorToMove) {
			fmt.Println("Check!")
		}
		if b.colorToMove == White {
			fmt.Print("White> ")
		} else {
			fmt.Print("Black> ")
		}
		if !scanner.Scan() {
			return scanner.Err()
		}
		s := strings.TrimSpace(scanner.Text())
		if s == "quit" {
			return nil
		}
		if s == "draw" {
			if drawOffer == opposite(b.colorToMove) {
				fmt.Println("Draw by agreement.")
				return nil
			}
			drawOffer = b.colorToMove
			fmt.Println("Draw offered; play your move. Opponent can accept with draw.")
			continue
		}
		if s == "resign" {
			if b.colorToMove == White {
				fmt.Println("White resigns. Black wins.")
			} else {
				fmt.Println("Black resigns. White wins.")
			}
			return nil
		}
		words := strings.Fields(s)
		if len(words) > 0 && words[0] == "claim" {
			position, count := b, repetitions[b.key()]
			if len(words) == 2 {
				m, err := findMove(words[1], legal)
				if err != nil {
					fmt.Println(err)
					continue
				}
				position.makeMove(m)
				count = repetitions[position.key()] + 1
			} else if len(words) != 1 {
				fmt.Println("Use claim or claim <move>.")
				continue
			}
			if result := drawClaim(position, count); result != "" {
				fmt.Println(result)
				return nil
			}
			fmt.Println("No threefold repetition or fifty-move draw available.")
			continue
		}
		var m Move
		if len(words) > 0 && words[0] == "engine" {
			depth := 6
			var depthErr error
			if len(words) == 2 {
				depth, depthErr = strconv.Atoi(words[1])
			}
			if len(words) > 2 || depthErr != nil || depth < 1 || depth > 6 {
				fmt.Println("Use engine [depth], with a depth from 1 to 6.")
				continue
			}
			var ok bool
			m, ok = BestMove(b, depth)
			if !ok {
				fmt.Println("No engine move available.")
				continue
			}
			fmt.Println("Engine plays", m)
			err = nil
		} else {
			m, err = findMove(s, legal)
		}
		if err != nil {
			fmt.Println(err)
			continue
		}
		if drawOffer != b.colorToMove {
			drawOffer = None
		}
		b.makeMove(m)
		repetitions[b.key()]++
	}
}
