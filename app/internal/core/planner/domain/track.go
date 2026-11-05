package domain

import monster "app/internal/core/monster/domain"

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
			for i > 0 && p.Rows[i-1].Kind == FoodRow {
				i--
			}
			return Passed{Rows: p.Rows[:i], Step: r.Step, Month: r.Month}, true
		}
		weeks--
	}
	return Passed{}, false
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
	Started bool
	Date    int
	Who     Who
}

func (t Tracking) Weeks(c Current) (weeks int, reset bool) {
	if !t.Started {
		return 0, false
	}
	if !c.Dated || c.Who != t.Who || c.Date < t.Date {
		return 0, true
	}
	return c.Date - t.Date, false
}
