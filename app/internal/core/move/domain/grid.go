package domain

import monster "app/internal/core/monster/domain"

const (
	// Columns and Rows are the size of a genus's grid of moves.
	Columns = 4
	Rows    = 6
)

// Genus names a breed.
type Genus struct {
	Code monster.Genus
	Name string
}

// Cell is one place in a genus's grid, with what the monster has done with the move in it.
type Cell struct {
	Move     *Move // nil for an empty cell
	Learned  bool
	CanLearn bool // not learned yet, and the monster meets the unlock rule; CurrentGrid makes it what the errantry teaches now
	Used     int
	Active   bool // the move is the active one of its column
	// Requirements is what the move still asks, condition by condition; it is empty for a move that is learned
	// or has no condition.
	Requirements []Requirement
}

// Grid is a genus's moves laid out as the game shows them, for one monster.
type Grid struct {
	Genus  Genus
	Cells  [Columns][Rows]Cell // Cells[col-1][row-1]
	Active [Columns]*Move      // the active move of each column; nil when it has none
}

// NewGrid lays a genus's moves out for a monster. A move whose cell or memory slot is outside the grid
// cannot be shown and is left out.
func NewGrid(genus Genus, moves []Move, m monster.Monster) Grid {
	return NewGridNamed(genus, moves, m, Names{})
}

// NewGridNamed is NewGrid with names for the genera in the unlock rules; the moves are named from the moves.
func NewGridNamed(genus Genus, moves []Move, m monster.Monster, names Names) Grid {
	g := Grid{Genus: genus}
	bySlot := make(map[int]*Move, len(moves))
	placeable := func(mv *Move) bool {
		return mv.Col >= 1 && mv.Col <= Columns && mv.Row >= 1 && mv.Row <= Rows && mv.MemorySlot >= 0 && mv.MemorySlot < monster.SlotCount
	}
	for i := range moves {
		if mv := &moves[i]; placeable(mv) {
			bySlot[mv.MemorySlot] = mv
		}
	}
	if names.Move == nil {
		names.Move = func(slot int) string {
			if mv, ok := bySlot[slot]; ok {
				return mv.Name
			}
			return ""
		}
	}
	for i := range moves {
		mv := &moves[i]
		if !placeable(mv) {
			continue
		}
		learned := m.Learned[mv.MemorySlot]
		cell := Cell{
			Move:     mv,
			Learned:  learned,
			CanLearn: !learned && mv.Unlock.Holds(m),
			Used:     m.Used[mv.MemorySlot],
		}
		if !learned {
			cell.Requirements = mv.Unlock.Requirements(m, names)
		}
		g.Cells[mv.Col-1][mv.Row-1] = cell
	}
	// Column c's active move must be one of that column's six memory slots, as the game keeps them.
	for c := range Columns {
		slot := m.Active[c]
		if slot < c*Rows || slot >= (c+1)*Rows {
			continue
		}
		mv, ok := bySlot[slot]
		if !ok {
			continue
		}
		g.Active[c] = mv
		g.Cells[mv.Col-1][mv.Row-1].Active = true
	}
	return g
}
