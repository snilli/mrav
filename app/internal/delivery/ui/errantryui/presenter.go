package errantryui

import (
	"fmt"
	"math"

	monster "app/internal/core/monster/domain"
	movedomain "app/internal/core/move/domain"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
)

type Ready struct {
	Cell   movedomain.Cell
	Chance profileui.Box
}

type RowView struct {
	Place string
	Stat  monster.Stat
	Gains []profileui.Segment
	Ready []Ready
	Weeks [domain.ErrantryWeeks]profileui.Box
}

type View struct {
	Rows []RowView
}

func NewView(p domain.Profile, g movedomain.Grid) View {
	m := p.Monster
	cells := map[int]movedomain.Cell{}
	var moves []movedomain.Move
	for _, column := range g.Cells {
		for _, cell := range column {
			if cell.Move != nil {
				cells[cell.Move.MemorySlot] = cell
				moves = append(moves, *cell.Move)
			}
		}
	}
	learn := domain.LearnChances(m, moves)

	var v View
	for i, o := range domain.PlanErrantry(m) {
		row := RowView{
			Place: o.Locale.Name, Stat: o.Locale.Stat,
			Gains: gainSegments(o), Ready: readyMoves(learn[i], cells),
		}
		for week, chance := range o.Chances {
			row.Weeks[week] = profileui.Box{Text: fmt.Sprintf("%.0f%%", chance), Tone: chanceTone(chance)}
		}
		v.Rows = append(v.Rows, row)
	}
	return v
}

func gainSegments(o domain.ErrantryOutlook) []profileui.Segment {
	stat, life := o.Locale.Stat, monster.Life
	return []profileui.Segment{
		{Text: profileui.StatNames[stat], Color: profileui.StatColors[stat]},
		{Text: fmt.Sprintf(" +%d-%d", o.Gains.Min, o.Gains.Max)},
		{Text: " · "},
		{Text: profileui.StatNames[life], Color: profileui.StatColors[life]},
		{Text: fmt.Sprintf(" +%d", o.Gains.Life)},
	}
}

func readyMoves(learn []domain.Learn, cells map[int]movedomain.Cell) []Ready {
	var ready []Ready
	for _, l := range learn {
		if l.Shown() {
			ready = append(ready, Ready{
				Cell:   cells[l.Move.MemorySlot],
				Chance: profileui.Box{Text: profileui.ChanceText(l.Chance), Tone: chanceTone(roundTenth(l.Chance))},
			})
		}
	}
	return ready
}

func roundTenth(percent float64) float64 { return math.Round(percent*10) / 10 }

func chanceTone(percent float64) profileui.Tone {
	switch {
	case percent >= 100:
		return profileui.ToneLimeGreen
	case percent >= 60:
		return profileui.ToneLightGreen
	case percent >= 30:
		return profileui.ToneLightYellow
	case percent > 0:
		return profileui.ToneOrange
	}
	return profileui.ToneOrangeRed
}
