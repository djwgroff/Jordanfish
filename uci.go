package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func runUCI(input io.Reader, output io.Writer) error {
	b, err := boardFromFEN(startFEN)
	if err != nil {
		return err
	}
	depth := 3
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "uci":
			fmt.Fprintln(output, "id name Jordanfish")
			fmt.Fprintln(output, "option name Depth type spin default 3 min 1 max 64")
			fmt.Fprintln(output, "uciok")
		case "isready":
			fmt.Fprintln(output, "readyok")
		case "ucinewgame":
			b, _ = boardFromFEN(startFEN)
		case "setoption":
			if len(words) >= 3 && words[1] == "name" && strings.EqualFold(words[2], "Depth") {
				if len(words) != 5 || words[3] != "value" {
					fmt.Fprintln(output, "info string use setoption name Depth value N")
					continue
				}
				value, err := parseUCIDepth(words[4])
				if err != nil {
					fmt.Fprintln(output, "info string", err)
					continue
				}
				depth = value
			}
		case "position":
			next, err := parseUCIPosition(words[1:])
			if err != nil {
				fmt.Fprintln(output, "info string", err)
				continue
			}
			b = next
		case "go":
			searchDepth := depth
			var depthErr error
			for i := 1; i < len(words); i++ {
				if words[i] == "depth" {
					if i+1 == len(words) {
						depthErr = fmt.Errorf("missing search depth")
						break
					}
					searchDepth, depthErr = parseUCIDepth(words[i+1])
					if depthErr != nil {
						break
					}
					i++
				}
			}
			if depthErr != nil {
				fmt.Fprintln(output, "info string", depthErr)
				continue
			}
			fmt.Fprintln(output, "info depth", searchDepth)
			move, ok := BestMove(b, searchDepth)
			if ok {
				fmt.Fprintln(output, "bestmove", move)
			} else {
				fmt.Fprintln(output, "bestmove 0000")
			}
		case "quit":
			return nil
		}
	}
	return scanner.Err()
}

func parseUCIDepth(s string) (int, error) {
	depth, err := strconv.Atoi(s)
	if err != nil || depth < 1 || depth > 64 {
		return 0, fmt.Errorf("depth must be from 1 to 64")
	}
	return depth, nil
}

func parseUCIPosition(words []string) (Board, error) {
	if len(words) == 0 {
		return Board{}, fmt.Errorf("missing position")
	}
	fen, rest := startFEN, words[1:]
	switch words[0] {
	case "startpos":
	case "fen":
		if len(words) < 7 {
			return Board{}, fmt.Errorf("position fen needs all six fields")
		}
		fen, rest = strings.Join(words[1:7], " "), words[7:]
	default:
		return Board{}, fmt.Errorf("use position startpos or position fen")
	}
	b, err := boardFromFEN(fen)
	if err != nil {
		return Board{}, err
	}
	if len(rest) > 0 {
		if rest[0] != "moves" {
			return Board{}, fmt.Errorf("expected moves after position")
		}
		for _, s := range rest[1:] {
			move, err := findMove(s, GenerateLegalMoves(b))
			if err != nil {
				return Board{}, fmt.Errorf("invalid position move %s", s)
			}
			b.makeMove(move)
		}
	}
	return b, nil
}
