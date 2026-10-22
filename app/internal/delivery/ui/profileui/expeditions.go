package profileui

import (
	"cmp"
	"fmt"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
)

type ExpeditionsView struct {
	Trips []TripView
	None  string
	Note  string
}

type TripView struct {
	Name       string
	When       string
	Conditions []ConditionView
}

type ConditionView struct {
	Text string
	Met  bool
}

func newExpeditionsView(p domain.Profile) ExpeditionsView {
	r := p.Player.Ranch
	var v ExpeditionsView
	list, showAll := domain.Expeditions, r.PhoenixKnown
	if domain.PhoenixPending(r) {
		list = []domain.Expedition{domain.Phoenix}
		v.Note = "The other expeditions open after this one"
	}
	type trip struct {
		e     domain.Expedition
		weeks int
	}
	trips := make([]trip, len(list))
	for i, e := range list {
		trips[i] = trip{e, r.WeeksUntil(e.Month, e.Week)}
	}
	if r.Known {
		slices.SortStableFunc(trips, func(a, b trip) int { return cmp.Compare(a.weeks, b.weeks) })
	}

	for _, t := range trips {
		if !showAll && !t.e.Fit(p.Monster) {
			continue
		}
		tv := TripView{Name: t.e.Name, When: fmt.Sprintf("%s wk %d", player.MonthNames[t.e.Month-1], t.e.Week)}
		if t.e.Host != "" {
			tv.Name += " (" + t.e.Host + ")"
		}
		if r.Known {
			tv.When += ", " + inWeeks(t.weeks)
		}
		for _, c := range t.e.Conditions(p.Monster) {
			tv.Conditions = append(tv.Conditions, conditionView(c))
		}
		v.Trips = append(v.Trips, tv)
	}
	if len(v.Trips) == 0 {
		v.None = "No expedition is open to this monster yet"
	}
	return v
}

func conditionView(c domain.Condition) ConditionView {
	var text string
	switch c.Kind {
	case domain.CondLife:
		text = fmt.Sprintf("%s ≥ %d (%d)", StatNames[monster.Life], c.Need, c.Have)
	case domain.CondFame:
		text = fmt.Sprintf("Fame ≥ %d (%d)", c.Need, c.Have)
	case domain.CondGrade:
		text = fmt.Sprintf("Grade ≥ %s (%s)", domain.RankLetter(c.Need), domain.RankLetter(c.Have))
	}
	return ConditionView{Text: text, Met: c.Met()}
}

func newNeedRow() needRow {
	n := needRow{mark: canvas.NewText("✗", toneColors[ToneOrangeRed]), label: widget.NewLabel("")}
	n.label.Wrapping = fyne.TextWrapWord
	n.object = container.NewBorder(nil, nil, n.mark, nil, n.label)
	return n
}

func (n needRow) set(c ConditionView) {
	n.mark.Text, n.mark.Color = "✗", toneColors[ToneOrangeRed]
	if c.Met {
		n.mark.Text, n.mark.Color = "✓", toneColors[ToneLimeGreen]
	}
	n.mark.Refresh()
	n.label.SetText(c.Text)
}
