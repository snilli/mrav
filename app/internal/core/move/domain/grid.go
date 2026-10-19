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
	CanLearn bool // not learned yet, and the monster meets the unlock rule
	Used     int
	Active   bool // the move is the active one of its column
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
	g := Grid{Genus: genus}
	bySlot := make(map[int]*Move, len(moves))
	for i := range moves {
		mv := &moves[i]
		if mv.Col < 1 || mv.Col > Columns || mv.Row < 1 || mv.Row > Rows || mv.MemorySlot < 0 || mv.MemorySlot >= monster.SlotCount {
			continue
		}
		bySlot[mv.MemorySlot] = mv
		learned := m.Learned[mv.MemorySlot]
		g.Cells[mv.Col-1][mv.Row-1] = Cell{
			Move:     mv,
			Learned:  learned,
			CanLearn: !learned && mv.Unlock.Holds(m),
			Used:     m.Used[mv.MemorySlot],
		}
	}
	// Column c's active move must be one of that column's six memory slots, as the game keeps them.
	for c := 0; c < Columns; c++ {
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
