// Package routeui is the Route panel: for each map of an expedition, a table of its tiles with the best route for
// the monster marked on it, and the places to search in the order to go.
package routeui

import (
	"fmt"
	"math"
	"slices"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
)

// View is a map with the route to take on it, and what to say about it.
type View struct {
	Name         string
	Summary      string        // what the monster brings to it, and what the route costs and is worth
	Stops        []string      // the places to search, numbered in the order to go
	Intelligence int           // the monster's, which says which gates it can find the way through
	Marker       *domain.Point // the tile the monster stands on, when it is on the trip of this map; nil otherwise
	Map          domain.ExpeditionMap
	Route        domain.Route
}

// NewView plans the best route on a map for a monster: with the energy that its life gives it to begin with, the
// chance that its power gives it of breaking a rock, and the intelligence that finds a way through a gate. What a
// place is worth is what a search of it turns up on average for the stat that the building is searched by, and for
// the nature of the monster, which the game takes as good above 0.
func NewView(p domain.Profile, m domain.ExpeditionMap) (View, error) {
	mon := p.Monster
	walker := domain.Walker{
		Energy:       domain.StartEnergy(mon.Stats[monster.Life]),
		BreakChance:  domain.BreakChance(mon.Stats[monster.Pow]),
		Intelligence: mon.Stats[monster.Int],
	}
	good := mon.Nature > 0
	m.Spots = slices.Clone(m.Spots) // the map is the catalog's: its places are not to be given what they are worth for this monster
	for i, s := range m.Spots {
		if s.Loot != nil {
			m.Spots[i].Value = s.Loot.Worth(mon.Stats[s.Loot.Stat], good)
		}
	}

	// On its trip, where the monster stands and what it has left are what a route is from, once it has made steps; the
	// energy that the game keeps from another trip, or from a longer-lived monster, is no more than a mark, so
	// more than it begins with is not taken for what it has left.
	var marker *domain.Point
	left := false
	ranch := p.Player.Ranch
	if m.Here(ranch) {
		x, y := ranch.Spot()
		marker = &domain.Point{X: x, Y: y}
		if left = ranch.EnergyKnown && ranch.Steps > 0 && ranch.Energy <= walker.Energy; left {
			walker.Energy, walker.At = ranch.Energy, marker
		}
	}

	route, err := domain.PlanRoute(m, walker)
	if err != nil {
		return View{}, err
	}
	nature := "bad"
	if good {
		nature = "good"
	}
	v := View{Name: m.Name, Intelligence: walker.Intelligence, Marker: marker, Map: m, Route: route}
	head := fmt.Sprintf("%d energy, ", walker.Energy)
	if left {
		head = fmt.Sprintf("%d energy left, from x %d, y %d, ", walker.Energy, marker.X, marker.Y)
	}
	head += fmt.Sprintf("rocks break %d%%, INT %d, %s nature: ", round(walker.BreakChance*100), walker.Intelligence, nature)
	if len(route.Spots) == 0 {
		v.Summary = head + "no place is in reach"
		return v, nil
	}
	v.Summary = head + fmt.Sprintf("%s, about %d energy, worth about %s",
		places(len(route.Spots)), round(route.Cost), profileui.Gold(round(route.Value)))
	for i, s := range route.Spots {
		name := s.Name
		if s.Loot != nil && s.Loot.Building != "" {
			name += " · " + s.Loot.Building
		}
		v.Stops = append(v.Stops, fmt.Sprintf("%d. %s · %s", i+1, name, profileui.Gold(round(s.Value))))
	}
	return v, nil
}

// Title is the line that the panel of a view has when it is closed.
func (v View) Title() string {
	if len(v.Route.Spots) == 0 {
		return fmt.Sprintf("Route · %s: nothing in reach", v.Name)
	}
	return fmt.Sprintf("Route · %s: %s, about %s", v.Name, places(len(v.Route.Spots)), profileui.Gold(round(v.Route.Value)))
}

func round(f float64) int { return int(math.Round(f)) }

func places(n int) string {
	if n == 1 {
		return "1 place"
	}
	return fmt.Sprintf("%d places", n)
}
