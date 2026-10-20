package profileui

import (
	"cmp"
	"fmt"
	"slices"

	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
)

// ExpeditionsView is what the expeditions panel shows: the trips that the monster is fit for, nearest first when the
// date is known. A trip that can be made once only is left out, because there is no telling whether it was made.
type ExpeditionsView struct {
	Trips []TripView
	None  string // what to say when there is no trip to show
}

// TripView is one expedition.
type TripView struct {
	Name string
	When string // the month and week, and how long to wait for it where the date is known
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

	var v ExpeditionsView
	for _, t := range trips {
		if !t.e.Fit(p.Monster) {
			continue
		}
		tv := TripView{Name: t.e.Name, When: fmt.Sprintf("%s wk %d", player.MonthNames[t.e.Month-1], t.e.Week)}
		if t.e.Host != "" {
			tv.Name += " (" + t.e.Host + ")"
		}
		if r.Known {
			tv.When += ", " + inWeeks(t.weeks)
		}
		v.Trips = append(v.Trips, tv)
	}
	if len(v.Trips) == 0 {
		v.None = "No expedition is open to this monster yet"
	}
	return v
}
