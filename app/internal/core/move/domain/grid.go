package domain

import monster "app/internal/core/monster/domain"

const (
	Columns = 4
	Rows    = 6
)

type Genus struct {
	Code monster.Genus
	Name string
}

type Cell struct {
	Move         *Move
	Learned      bool
	CanLearn     bool
	Chance       float64
	Place        string
	Used         int
	Active       bool
	Requirements []Requirement
}

type Grid struct {
	Genus  Genus
	Cells  [Columns][Rows]Cell
	Active [Columns]*Move
}

func NewGrid(genus Genus, moves []Move, m monster.Monster) Grid {
	return NewGridNamed(genus, moves, m, Names{})
}

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
