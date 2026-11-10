package domain

import (
	"slices"

	monster "app/internal/core/monster/domain"
)

type Passed struct {
	Rows  []Row
	Step  int
	Month int
}

func (p Plan) Passed(weeks int) (Passed, bool) {
	for i, r := range p.Rows {
		if r.Kind == FoodRow {
			continue
		}
		if weeks == 0 || r.Kind == NextRow {
			if r.Kind == FollowRow {
				return Passed{}, false
			}
			return Passed{Rows: p.Rows[:i], Step: r.Step, Month: r.Month}, true
		}
		weeks--
	}
	return Passed{}, false
}

func (p Passed) Played(itemUsed bool, after Settings) Passed {
	p.Rows = slices.Clone(p.Rows)
	weeks := make([]int, 0, len(p.Rows))
	for i, r := range p.Rows {
		if r.Kind != FoodRow {
			weeks = append(weeks, i)
		}
	}
	if len(weeks) == 0 {
		return p
	}
	if !itemUsed {
		p.Rows[weeks[0]].Item = NoItem
	}
	last := &p.Rows[weeks[len(weeks)-1]]
	last.Fatigue, last.Stress, last.LI, last.Form = after.Fatigue, after.Stress, lifeIndex(after.Fatigue, after.Stress), after.Form
	return p
}

func (r Row) Aged(start, lifespan int) (Row, bool) {
	if lifespan >= start {
		return r, false
	}
	if r.ShowLife {
		drift := lifespan - r.Life
		r.Life, r.Lost, r.Total = lifespan, r.Lost+drift, r.Total+drift
	}
	return r, true
}

type Who struct {
	Main, Sub monster.Genus
	Name      string
	Lifespan  int
}

func WhoOf(m monster.Monster) Who {
	return Who{Main: m.Genus, Sub: m.SubGenus, Name: m.Name, Lifespan: m.InitialLifespan}
}

type Tracking struct {
	Started     bool
	Start, Date int
	Who         Who
}

func (t Tracking) Weeks(c Current) (weeks int, ok bool) {
	if !t.Started || !c.Dated || c.Who != t.Who || c.Date < t.Start {
		return 0, false
	}
	return c.Date - t.Date, true
}

func (t Tracking) Another(c Current) bool { return t.Started && c.Dated && c.Who != t.Who }
