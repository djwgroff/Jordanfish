package main

const tableCapacity = 1 << 16

type scoreBound byte

const (
	exactScore scoreBound = iota
	lowerBound
	upperBound
)

type tableEntry struct {
	depth int
	score int
	bound scoreBound
	move  Move
}

type transpositionTable map[Board]tableEntry

func tableKey(b Board) Board {
	b.fullmove = 0
	return b
}

// mate distances need to be relative to the cached position rather than the root
func scoreToTable(score, ply int) int {
	if score >= mateScore-1024 {
		return score + ply
	}
	if score <= -mateScore+1024 {
		return score - ply
	}
	return score
}

func scoreFromTable(score, ply int) int {
	return scoreToTable(score, -ply)
}

func (table transpositionTable) store(key Board, entry tableEntry) {
	if table == nil {
		return
	}
	if old, exists := table[key]; exists {
		if old.depth > entry.depth {
			return
		}
	} else if len(table) >= tableCapacity {
		return
	}
	table[key] = entry
}
