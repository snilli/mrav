// Package errantryui is the Errantry panel: for each place that a monster can be sent to, the stat that the trip trains
// and what it would give, the techs that the monster could learn there, and the chance that each of the four weeks
// succeeds, for the monster as it is now.
package errantryui

import (
	"fmt"
	"math"
	"strconv"

	monster "app/internal/core/monster/domain"
	movedomain "app/internal/core/move/domain"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
)

// Ready is a move that a place may teach the monster: the cell of the grid that holds it, which says where it stands and what it
// needs, and the chance that the place teaches it.
type Ready struct {
	Cell   movedomain.Cell
	Chance profileui.Box
}

// RowView is one place: its name, the stat that the trip trains and what it would give, the moves that it may teach the monster
// as the place tries them, and the chance of each week.
type RowView struct {
	Place string
	Stat  monster.Stat
	Gains []profileui.Segment
	Ready []Ready
	Weeks [domain.ErrantryWeeks]profileui.Box
}

// View is what the panel shows.
type View struct {
	Rows []RowView
}

func NewView(p domain.Profile, g movedomain.Grid) View {
	m := p.Monster
	cells := map[int]movedomain.Cell{} // by memory slot
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

// gainSegments say what a trip would give: the stat of the place from its least to its most, and the Life, each stat named in
// its colour.
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

// readyMoves are the moves that a place may teach, with their chances, as the place tries them. A move whose chance is nothing to
// one decimal is left out: the calculator of the page writes it as 0%.
func readyMoves(learn []domain.Learn, cells map[int]movedomain.Cell) []Ready {
	var ready []Ready
	for _, l := range learn {
		if l.Shown() {
			ready = append(ready, Ready{
				Cell:   cells[l.Move.MemorySlot],
				Chance: profileui.Box{Text: chanceText(l.Chance), Tone: chanceTone(roundTenth(l.Chance))},
			})
		}
	}
	return ready
}

// roundTenth is a percent to one decimal, as the calculator of the page rounds it.
func roundTenth(percent float64) float64 { return math.Round(percent*10) / 10 }

// chanceText says a chance in percent as the calculator of the page does: to one decimal, and without the decimal when it is 0.
func chanceText(percent float64) string {
	return strconv.FormatFloat(roundTenth(percent), 'f', -1, 64) + "%"
}

// chanceTone is the tone of a chance in percent, warmer as it falls: green when sure or likely, yellow when even, orange
// when unlikely, and red when none.
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
