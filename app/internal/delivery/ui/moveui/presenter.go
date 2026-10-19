// Package moveui is the Move Viewer window: a pure presenter that turns a grid into what the screen
// shows, and the Fyne widgets that draw it.
package moveui

import (
	"strconv"

	"app/internal/core/move/domain"
)

// The names of the overlays in the icon assets.
const (
	NoMove    = "No_Move"
	Checked   = "skillchecked"
	CanGet    = "skillget"
	Highlight = "hilight"
)

// emptyText is what the game says about an empty cell.
const emptyText = "Click on a move to learn more."

// CellView is what one cell of the grid shows.
type CellView struct {
	Empty    bool
	Icon     string // an asset key; NoMove for an empty cell
	Learned  bool
	CanLearn bool
	Active   bool
}

// GridView is what the whole grid shows.
type GridView struct {
	Genus  string
	Cells  [domain.Columns][domain.Rows]CellView // Cells[col-1][row-1]
	Active [domain.Columns]string                // the icon of each column's active move; NoMove when it has none
}

// RequirementView is one thing a move asks of a monster before it can learn it, and where the monster stands on it.
type RequirementView struct {
	Text string
	Have string // what the monster has, when that is a number; empty otherwise
	Met  bool
}

// DetailView is what the details of one cell show, every value already in the game's words.
type DetailView struct {
	Name        string
	Kind        domain.Kind
	Guts        string
	Damage      string
	Hit         string
	Withering   string
	Sharpness   string
	Extra       string // "Recovery: B (30)"; empty when the move has no extra row
	Uses        string
	Learned     bool
	CanLearn    bool
	Description string
	// Requirements are what the move still asks of the monster; none for a learned move or an empty cell.
	Requirements []RequirementView
}

func NewGridView(g domain.Grid) GridView {
	v := GridView{Genus: g.Genus.Name}
	for c := range g.Cells {
		for r, cell := range g.Cells[c] {
			v.Cells[c][r] = newCellView(cell)
		}
		v.Active[c] = NoMove
		if m := g.Active[c]; m != nil {
			v.Active[c] = m.Icon
		}
	}
	return v
}

func newCellView(c domain.Cell) CellView {
	if c.Move == nil {
		return CellView{Empty: true, Icon: NoMove}
	}
	return CellView{Icon: c.Move.Icon, Learned: c.Learned, CanLearn: c.CanLearn, Active: c.Active}
}

func NewDetailView(c domain.Cell) DetailView {
	m := c.Move
	if m == nil {
		return DetailView{
			Name: "No Move", Kind: domain.KindUnknown, Guts: "---", Damage: "---", Hit: "---",
			Withering: "---", Sharpness: "---", Uses: "---", Description: emptyText,
		}
	}
	v := DetailView{
		Name:        m.Name,
		Kind:        m.Kind,
		Guts:        strconv.Itoa(m.Guts),
		Damage:      domain.Rate(domain.ScaleForce, m.Force),
		Hit:         domain.Rate(domain.ScaleHit, m.Hit),
		Withering:   domain.Rate(domain.ScaleWither, m.Wither),
		Sharpness:   domain.Rate(domain.ScaleCrit, m.Crit),
		Uses:        strconv.Itoa(c.Used),
		Learned:     c.Learned,
		CanLearn:    c.CanLearn,
		Description: m.Description,
	}
	for _, r := range c.Requirements {
		v.Requirements = append(v.Requirements, RequirementView{Text: r.Text, Have: r.Have, Met: r.Met})
	}
	if e := m.Extra; e != nil {
		value := e.Text
		if e.Scale != nil {
			value = domain.Rate(*e.Scale, e.Raw)
		}
		v.Extra = e.Label + " " + value
	}
	return v
}
