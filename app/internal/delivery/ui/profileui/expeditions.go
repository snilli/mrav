package profileui

import (
	"cmp"
	"fmt"
	"slices"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
)

// ExpeditionsView is what the expeditions panel shows: the trips the monster is invited on, nearest first when the
// date is known, what each asks of the monster, and the energy it would start with and the game now has.
type ExpeditionsView struct {
	Title       string // the nearest trip, and whether the monster is fit for it
	StartEnergy string
	GameEnergy  string // what the game last kept as an expedition's energy; empty where it could not be read
	GameWalk    string // the tile of the map it last stood on and the steps it had walked; empty where not read
	Costs       string
	Trips       []TripView
}

// TripView is one expedition.
type TripView struct {
	Name       string
	When       string // the month and week, and how long to wait for it where the date is known
	Note       string
	Conditions []ConditionView
	Fit        bool // the monster meets every condition
}

// ConditionView is one thing an invitation asks of the monster, and where the monster stands on it.
type ConditionView struct {
	Text string
	Have string
	Met  bool
}

func newExpeditionsView(p domain.Profile) ExpeditionsView {
	r := p.Player.Ranch
	type trip struct {
		e     domain.Expedition
		weeks int
	}
	trips := make([]trip, len(domain.Expeditions))
	for i, e := range domain.Expeditions {
		trips[i] = trip{e, r.WeeksUntil(e.Month, e.Week)}
	}
	if r.Known { // the nearest first; the calendar's order is the order of the ties and of a date not known
		slices.SortStableFunc(trips, func(a, b trip) int { return cmp.Compare(a.weeks, b.weeks) })
	}

	v := ExpeditionsView{
		StartEnergy: fmt.Sprintf("Starts with %d energy (70 for every 70 LIF, from 140 up to 700)", domain.StartEnergy(p.Monster.Stats[monster.Life])),
		Costs: fmt.Sprintf("A step costs %d energy, an obstacle %d, a search %d; ending under %d costs %d weeks of life",
			domain.MoveCost, domain.ObstacleCost, domain.SearchCost, domain.FailBelow, domain.FailWeeks),
	}
	if r.EnergyKnown {
		v.GameEnergy = fmt.Sprintf("Energy in the game now: %d (it stays after an expedition)", r.Energy)
	}
	if r.WalkKnown {
		x, y := r.Spot()
		v.GameWalk = fmt.Sprintf("Position in the game now: x %d, y %d · %d step", x, y, r.Steps)
		if r.Steps != 1 {
			v.GameWalk += "s"
		}
		v.GameWalk += " walked"
	}
	for _, t := range trips {
		tv := TripView{Name: t.e.Name, When: fmt.Sprintf("%s wk %d", player.MonthNames[t.e.Month-1], t.e.Week), Fit: true}
		if t.e.Host != "" {
			tv.Name += " (" + t.e.Host + ")"
		}
		if r.Known {
			tv.When += ", " + inWeeks(t.weeks)
		}
		if t.e.First {
			tv.Note = "the first expedition: made once, and it unlocks the others"
		}
		for _, c := range t.e.Conditions(p.Monster) {
			tv.Conditions = append(tv.Conditions, ConditionView{Text: c.Text, Have: c.Have, Met: c.Met})
			tv.Fit = tv.Fit && c.Met
		}
		v.Trips = append(v.Trips, tv)
	}
	if len(v.Trips) > 0 {
		next := v.Trips[0]
		v.Title = "Next expedition: " + next.Name + ", " + next.When
		if next.Fit {
			v.Title += " · fit"
		} else {
			v.Title += " · not fit yet"
		}
	}
	return v
}
