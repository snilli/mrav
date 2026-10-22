package moveui

import (
	"strconv"
	"strings"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
	"app/internal/delivery/ui/profileui"
)

const (
	NoMove    = "No_Move"
	Highlight = "hilight"
)

const (
	statusLearned = "LEARNED"
	statusReady   = "READY TO LEARN"
	statusLocked  = "NOT YET"
)

const emptyText = "Click on a move to learn more."

type CellView struct {
	Empty    bool
	Icon     string
	Learned  bool
	CanLearn bool
	Chance   string
	Locked   bool
	Active   bool
	Kind     domain.Kind
	Nature   domain.Nature
}

type GridView struct {
	Genus                          string
	Cells                          [domain.Columns][domain.Rows]CellView
	Active                         [domain.Columns]string
	ActiveKind                     [domain.Columns]domain.Kind
	Learned, Ready, Chance, Locked int
}

type RequirementView struct {
	Text  string
	Have  string
	Met   bool
	Stats []monster.Stat
}

type DetailView struct {
	Name         string
	Kind         domain.Kind
	Guts         string
	Damage       string
	Hit          string
	Withering    string
	Sharpness    string
	Extra        string
	Uses         string
	Learned      bool
	CanLearn     bool
	Chance       string
	Status       string
	Nature       domain.Nature
	Description  string
	Requirements []RequirementView
}

func NewGridView(g domain.Grid) GridView {
	v := GridView{Genus: g.Genus.Name}
	for c := range g.Cells {
		for r, cell := range g.Cells[c] {
			cv := newCellView(cell)
			v.Cells[c][r] = cv
			switch {
			case cv.Empty:
			case cv.Learned:
				v.Learned++
			case cv.CanLearn:
				v.Ready++
			case cv.Chance != "":
				v.Chance++
			default:
				v.Locked++
			}
		}
		v.Active[c] = NoMove
		if m := g.Active[c]; m != nil {
			v.Active[c], v.ActiveKind[c] = m.Icon, m.Kind
		}
	}
	return v
}

func newCellView(c domain.Cell) CellView {
	if c.Move == nil {
		return CellView{Empty: true, Icon: NoMove}
	}
	chance := chanceOf(c)
	return CellView{
		Icon: c.Move.Icon, Learned: c.Learned, CanLearn: c.CanLearn, Chance: chance, Locked: !c.Learned && !c.CanLearn && chance == "",
		Active: c.Active, Kind: c.Move.Kind, Nature: c.Move.Unlock.Nature(),
	}
}

func chanceOf(c domain.Cell) string {
	if c.Learned || c.CanLearn || c.Chance <= 0 {
		return ""
	}
	return profileui.ChanceText(c.Chance)
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
		Chance:      chanceOf(c),
		Nature:      m.Unlock.Nature(),
		Description: describe(kindOf(m.Description), c.Place),
	}
	switch {
	case c.Learned:
		v.Status = statusLearned
	case c.CanLearn:
		v.Status = statusReady
	case v.Chance != "":
		v.Status = v.Chance
	default:
		v.Status = statusLocked
	}
	for _, r := range c.Requirements {
		v.Requirements = append(v.Requirements, RequirementView{Text: r.Text, Have: r.Have, Met: r.Met, Stats: r.Stats})
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

func kindOf(description string) string {
	first, _, _ := strings.Cut(description, ".")
	return strings.TrimSuffix(first, " tech")
}

func describe(kind, place string) string {
	var parts []string
	if kind != "" {
		parts = append(parts, "Type: "+kind)
	}
	if place != "" {
		parts = append(parts, "Location: "+place)
	}
	return strings.Join(parts, " · ")
}
