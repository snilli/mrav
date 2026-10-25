package routeui

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	"fyne.io/fyne/v2/widget"

	"app/internal/core/profile/domain"
)

const findColumns = 2

type FindView struct {
	Items  []string
	Places []FindPlace
	NotYet []string
	Counts map[string]int
}

type FindPlace struct {
	Name     string
	Building string
	At       domain.Point
	Energy   string
	Items    []FindItem
}

type FindItem struct {
	Item   string
	Chance string
}

func NewFindView(v View, picked map[string]bool, by domain.Searcher) FindView {
	var f FindView
	m := v.Map
	seen := map[string]bool{}
	for _, s := range m.Spots {
		for _, c := range s.Chances {
			if !seen[c.Item] {
				seen[c.Item] = true
				f.Items = append(f.Items, c.Item)
			}
		}
	}
	slices.Sort(f.Items)

	type ranked struct {
		place FindPlace
		odds  float64
		cost  float64
		reach bool
	}
	var list []ranked
	for _, s := range m.Spots {
		r := ranked{place: FindPlace{Name: s.Name, Building: s.Building, At: s.At}}
		listed := map[string]bool{}
		for _, c := range s.Chances {
			if !picked[c.Item] || listed[c.Item] {
				continue
			}
			listed[c.Item] = true
			if pct, ok := lootPct(s, c.Item); ok {
				r.place.Items = append(r.place.Items, FindItem{Item: c.Item, Chance: chance(pct)})
				r.odds += pct
				continue
			}
			f.NotYet = append(f.NotYet, notYet(s, c.Item, by))
		}
		if len(r.place.Items) == 0 {
			continue
		}
		way, reach := v.Ways[s.Name]
		r.cost, r.reach = way.Cost, reach
		r.place.Energy = walkText(way, reach)
		list = append(list, r)
	}
	slices.SortStableFunc(list, func(a, b ranked) int {
		return cmp.Or(
			compareBool(b.reach, a.reach),
			cmp.Compare(len(b.place.Items), len(a.place.Items)),
			cmp.Compare(b.odds, a.odds),
			cmp.Compare(a.cost, b.cost),
		)
	})
	for _, r := range list {
		if f.Counts == nil {
			f.Counts = map[string]int{}
		}
		f.Counts[r.place.Name] = len(r.place.Items)
		f.Places = append(f.Places, r.place)
	}
	return f
}

func compareBool(a, b bool) int {
	switch {
	case a == b:
		return 0
	case a:
		return 1
	}
	return -1
}

func lootPct(s domain.Spot, item string) (float64, bool) {
	for _, d := range s.Loot {
		if d.Item == item {
			return d.Pct, true
		}
	}
	return 0, false
}

func notYet(s domain.Spot, item string, by domain.Searcher) string {
	where := s.Name
	if s.Building != "" {
		where += " (" + s.Building + ")"
	}
	stat := "INT"
	if s.BySkill {
		stat = "SKI"
	}
	need, ok := 0, false
	for _, c := range s.Chances {
		if c.Item == item && c.Good == by.GoodNature && c.Pct > 0 && (!ok || c.MinStat < need) {
			need, ok = c.MinStat, true
		}
	}
	if !ok {
		return fmt.Sprintf("%s at %s · not for this nature", item, where)
	}
	return fmt.Sprintf("%s at %s · needs %s %d", item, where, stat, need)
}

func (s *Screen) items() []string {
	var out []string
	for _, b := range s.boxes {
		for _, item := range NewFindView(b.v, nil, s.by).Items {
			if !slices.Contains(out, item) {
				out = append(out, item)
			}
		}
	}
	slices.Sort(out)
	return out
}

func (s *Screen) drawChecks() {
	want := strings.ToLower(s.filter.Text)
	all := s.items()
	open := s.listOpen || want != ""
	s.showItems.SetText("+ Add item")
	if open {
		s.showItems.SetText("Done")
	}
	s.chips.RemoveAll()
	for _, item := range all {
		if s.picked[item] {
			s.chips.Add(widget.NewButton(item+" ×", func() { s.pick(item, false) }))
		}
	}
	s.chips.Refresh()
	s.checks.Hidden = !open
	s.search.Hidden = !open
	s.checks.RemoveAll()
	for _, item := range all {
		if !strings.Contains(strings.ToLower(item), want) {
			continue
		}
		c := widget.NewCheck(item, nil)
		c.Checked = s.picked[item]
		c.OnChanged = func(on bool) { s.pick(item, on) }
		s.checks.Add(c)
	}
	s.checks.Refresh()
	s.picker.Refresh()
}

func (s *Screen) drawBoxes() {
	for _, b := range s.boxes {
		var f FindView
		if s.mode == modeFind {
			f = NewFindView(b.v, s.picked, s.by)
		}
		b.draw(s.mode, f, s.state)
	}
	s.list.Refresh()
}

func (s *Screen) pick(item string, on bool) {
	s.picked[item] = on
	s.drawChecks()
	s.drawBoxes()
}

func (s *Screen) pickAll(on bool) {
	for _, item := range s.items() {
		s.picked[item] = on
	}
	s.drawChecks()
	s.drawBoxes()
}
