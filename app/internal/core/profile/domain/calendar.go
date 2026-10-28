package domain

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	player "app/internal/core/player/domain"
)

type EventKind int

const (
	EventCup EventKind = iota
	EventInvite
	EventTrip
	EventRanch
	EventShop
	EventVisit
)

type Fit int

const (
	FitReady Fit = iota
	FitNotYet
	FitUnknown
	FitNever
	FitWarn
)

type CondType string

const (
	CondRankAtLeast        CondType = "rank_at_least"
	CondRankExactly        CondType = "rank_exactly"
	CondCupGrade           CondType = "cup_grade"
	CondTrainerRankAtLeast CondType = "trainer_rank_at_least"
	CondYearAtLeast        CondType = "year_at_least"
	CondAgeUnder           CondType = "age_under"
	CondAgeAtLeast         CondType = "age_at_least"
	CondNatureAtLeast      CondType = "nature_at_least"
	CondNatureAtMost       CondType = "nature_at_most"
	CondExpedition         CondType = "expedition"
	CondPhoenixOpen        CondType = "phoenix_open"
	CondPhoenixDone        CondType = "phoenix_done"
	CondIMaYear            CondType = "ima_year"
	CondFacilityNext       CondType = "facility_next"
)

type EntryCondition struct {
	Type  CondType
	Value int
}

type CalendarEntry struct {
	Month, Week int
	Name        string
	Kind        EventKind
	Purse       int
	Prize       string
	Unseen      string
	Note        string
	When        []EntryCondition
}

type Event struct {
	Name   string
	Kind   EventKind
	Fit    Fit
	Need   string
	Detail string
}

type CalendarWeek struct {
	Date   player.Date
	In     int
	Events []Event
}

const (
	ShopSale     = "Shop sale"
	ErrantrySale = "Errantry sale"
	imaEvery     = 4
)

func Calendar(p Profile) []CalendarWeek {
	r := p.Player.Ranch
	if !r.Known {
		return nil
	}
	last := r.Calendar().Year + 1
	var weeks []CalendarWeek
	for in := 0; ; in++ {
		at := r
		at.Date += in
		d := at.Calendar()
		if d.Year > last {
			return weeks
		}
		weeks = append(weeks, CalendarWeek{Date: d, In: in, Events: events(p, d, in)})
	}
}

func events(p Profile, d player.Date, in int) []Event {
	type ranked struct {
		Event
		cup   bool
		grade int
	}
	var list []ranked
	for _, e := range p.Calendar {
		if e.Month != d.Month || e.Week != d.Week {
			continue
		}
		if ev, ok := e.event(p, d, in); ok {
			list = append(list, ranked{ev, e.Kind == EventCup, e.grade()})
		}
	}
	slices.SortStableFunc(list, func(a, b ranked) int {
		switch {
		case a.cup && b.cup:
			return cmp.Compare(b.grade, a.grade)
		case a.cup:
			return -1
		case b.cup:
			return 1
		}
		return 0
	})
	var out []Event
	for _, r := range list {
		out = append(out, r.Event)
	}
	if in == 0 && p.Player.ErrantrySaleReady() {
		out = append(out, Event{Name: ErrantrySale, Kind: EventShop})
	}
	if in == p.Player.NextSale {
		out = append(out, Event{Name: ShopSale, Kind: EventShop})
	}
	return out
}

func (e CalendarEntry) grade() int {
	for _, c := range e.When {
		switch c.Type {
		case CondCupGrade, CondRankExactly, CondRankAtLeast:
			return c.Value
		}
	}
	return -1
}

func (e CalendarEntry) event(p Profile, d player.Date, in int) (Event, bool) {
	ev := Event{Name: e.Name, Kind: e.Kind, Detail: e.detail()}
	var needs []string
	for _, c := range e.When {
		fit, need, held := c.check(p, d, in, e.Name)
		if !held {
			return Event{}, false
		}
		switch fit {
		case FitNever:
			ev.Fit, ev.Need = FitNever, ""
			return ev, true
		case FitWarn:
			ev.Fit = FitWarn
			needs = append(needs, need)
		case FitNotYet:
			if ev.Fit != FitWarn {
				ev.Fit = FitNotYet
			}
			needs = append(needs, need)
		case FitUnknown:
			if ev.Fit == FitReady {
				ev.Fit = FitUnknown
			}
		}
	}
	switch {
	case ev.Fit == FitNotYet || ev.Fit == FitWarn:
		ev.Need = strings.Join(needs, ", ")
	case ev.Fit == FitReady && e.Unseen != "":
		ev.Fit, ev.Need = FitUnknown, e.Unseen
	case ev.Fit == FitReady:
		ev.Need = e.Note
	}
	return ev, true
}

func (e CalendarEntry) detail() string {
	var parts []string
	if e.Purse > 0 {
		parts = append(parts, fmt.Sprintf("%dG", e.Purse))
	}
	if e.Prize != "" {
		parts = append(parts, e.Prize)
	}
	return strings.Join(parts, ", ")
}

func (c EntryCondition) check(p Profile, d player.Date, in int, name string) (Fit, string, bool) {
	m, r := p.Monster, p.Player.Ranch
	age := m.Age + in
	v := c.Value
	switch c.Type {
	case CondRankAtLeast:
		return least(m.Rank >= v, "Rank "+RankLetter(v))
	case CondRankExactly:
		if m.Rank > v {
			return FitNever, "", true
		}
		return least(m.Rank >= v, "Rank "+RankLetter(v))
	case CondCupGrade:
		if m.Rank > v {
			return FitWarn, "Rank above " + RankLetter(v), true
		}
		return least(m.Rank >= v, "Rank "+RankLetter(v))
	case CondTrainerRankAtLeast:
		return least(r.Grade >= v, fmt.Sprintf("Trainer rank %d", v))
	case CondYearAtLeast:
		return least(d.Year >= v, fmt.Sprintf("Year %d", v))
	case CondAgeUnder:
		if age >= v {
			return FitNever, "", true
		}
		return FitReady, "", true
	case CondAgeAtLeast:
		return least(age >= v, fmt.Sprintf("Age %d weeks", v))
	case CondNatureAtLeast:
		return least(m.Nature >= v, fmt.Sprintf("Nature %d or more", v))
	case CondNatureAtMost:
		return least(m.Nature <= v, fmt.Sprintf("Nature %d or less", v))
	case CondExpedition:
		return expeditionFit(name, p)
	case CondPhoenixOpen:
		switch {
		case !r.PhoenixKnown:
			return FitUnknown, "", true
		case r.PhoenixDone:
			return FitNever, "", true
		}
		return FitReady, "", true
	case CondPhoenixDone:
		return least(!r.PhoenixKnown || r.PhoenixDone, "Phoenix first")
	case CondIMaYear:
		held := r.Tournament != 0 && d.Year >= r.Tournament && (d.Year-r.Tournament)%imaEvery == 0
		return FitReady, "", held
	case CondFacilityNext:
		up, ok := r.Next(player.Facility(v))
		if !ok {
			return FitNever, "", true
		}
		return least(r.Grade >= up.Grade, fmt.Sprintf("Trainer rank %d", up.Grade))
	}
	return FitUnknown, "", true
}

func least(met bool, need string) (Fit, string, bool) {
	if met {
		return FitReady, "", true
	}
	return FitNotYet, need, true
}

func expeditionFit(name string, p Profile) (Fit, string, bool) {
	for _, e := range append([]Expedition{Phoenix}, Expeditions...) {
		if e.Name != name {
			continue
		}
		var lack []string
		for _, c := range e.Conditions(p.Monster) {
			if c.Met() {
				continue
			}
			switch c.Kind {
			case CondLife:
				lack = append(lack, fmt.Sprintf("Life %d", c.Need))
			case CondFame:
				lack = append(lack, fmt.Sprintf("Fame %d", c.Need))
			case CondGrade:
				lack = append(lack, "Rank "+RankLetter(c.Need))
			}
		}
		if len(lack) > 0 {
			return FitNotYet, strings.Join(lack, ", "), true
		}
		return FitReady, "", true
	}
	return FitUnknown, "", true
}
