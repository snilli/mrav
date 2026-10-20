// Package routeui is the Route panel: for each map of an expedition, a table of its tiles with the best route for
// the monster marked on it, and the places to search in the order to go.
package routeui

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
)

// View is a map with the route to take on it, and what to say about it.
type View struct {
	Name         string
	Summary      string          // what the monster brings to it, and what the route costs and is worth
	Places       []PlaceView     // the places of the map: those that the route takes in the order to go, then the rest
	Teleports    []string        // where each teleport of the map leads, a line for each
	Intelligence int             // the monster's, which says which gates it can find the way through
	Marker       *domain.Point   // the tile the monster stands on, when it is on the trip of this map; nil otherwise
	Standing     []profileui.Box // what to say of the monster on the trip: its energy, where it is, its steps
	Map          domain.ExpeditionMap
	Route        domain.Route
}

// NewView plans the best route on a map for a monster: with the energy that its life gives it to begin with, the
// chance that its power gives it of breaking a rock, and the intelligence that finds a way through a gate. What a
// place is worth is in the map already: the database worked it out for this monster.
func NewView(p domain.Profile, m domain.ExpeditionMap) (View, error) {
	mon := p.Monster
	walker := domain.Walker{
		Energy:       domain.StartEnergy(mon.Stats[monster.Life]),
		BreakChance:  domain.BreakChance(mon.Stats[monster.Pow]),
		Intelligence: mon.Stats[monster.Int],
	}
	// what a place is worth is in the maps of the profile already, worked out for this monster, with its nature
	// On its trip, where the monster stands and what it has left are what a route is from, once it has made steps; the
	// energy that the game keeps from another trip, or from a longer-lived monster, is no more than a mark, so
	// more than it begins with is not taken for what it has left.
	var marker *domain.Point
	left := false
	ranch := p.Player.Ranch
	if m.Here(ranch) {
		at, _ := m.Stand(ranch)
		marker = &at
		if left = ranch.EnergyKnown && ranch.Steps > 0 && ranch.Energy <= walker.Energy; left {
			walker.Energy, walker.At = ranch.Energy, marker
		}
	}

	route, err := domain.PlanRoute(m, walker)
	if err != nil {
		return View{}, err
	}
	v := View{
		Name: m.Name, Intelligence: walker.Intelligence, Marker: marker, Map: m, Route: route,
		Places: placeViews(m, route), Teleports: teleports(m),
	}
	if marker != nil {
		v.Standing = standing(ranch)
	}
	// What the panel says elsewhere is not said again here: the energy that a monster on its trip has left and where it is are in
	// the boxes above the map, what the route is worth is in the heading, the INT is in the gates and the nature in the places.
	breaks := round(walker.BreakChance * 100)
	if left {
		v.Summary = fmt.Sprintf("Rocks break %d%%", breaks)
	} else {
		v.Summary = fmt.Sprintf("Starts with %d energy · rocks break %d%%", walker.Energy, breaks)
	}
	if len(route.Spots) > 0 {
		v.Summary += fmt.Sprintf(" · the route costs about %d energy", round(route.Cost))
	}
	return v, nil
}

// PlaceView is a place of the map, where it stands in the route, what a search of it is worth, and what it may turn up.
type PlaceView struct {
	Name     string // its letter on the map
	Building string
	Stop     int    // its number in the route, from 1; 0 where the route leaves it out
	Worth    string // what a search of it is worth on average, in words; empty where it is worth nothing
	Loot     []DropView
}

// DropView is an item that a search may turn up, and the chance that it does.
type DropView struct {
	Item   string
	Chance string // in percent, to one decimal, without a needless .0
	Story  bool   // an item of the story, which the game gives once
}

// storyItems are the items that the game gives once, which a guide marks on the maps: what the unlocks wait for.
var storyItems = map[string]bool{
	"Old Sheath": true, "Mask": true, "Big Footstep": true, "Fire Feather": true, "Strong Glue": true, "Gold Peach": true,
}

// placeViews lists the places of the map: the ones that the route takes, in the order to go, and then the ones that it
// leaves out, by name.
func placeViews(m domain.ExpeditionMap, route domain.Route) []PlaceView {
	stop := make(map[string]int, len(route.Spots))
	var out []PlaceView
	for i, s := range route.Spots {
		stop[s.Name] = i + 1
		out = append(out, placeView(s, i+1))
	}
	var rest []PlaceView
	for _, s := range m.Spots {
		if stop[s.Name] == 0 {
			rest = append(rest, placeView(s, 0))
		}
	}
	slices.SortStableFunc(rest, func(a, b PlaceView) int { return strings.Compare(a.Name, b.Name) })
	return append(out, rest...)
}

func placeView(s domain.Spot, stop int) PlaceView {
	p := PlaceView{Name: s.Name, Building: s.Building, Stop: stop}
	if round(s.Value) > 0 {
		p.Worth = "about " + profileui.Gold(round(s.Value))
	}
	for _, d := range s.Loot {
		p.Loot = append(p.Loot, DropView{Item: d.Item, Chance: chance(d.Pct), Story: storyItems[d.Item]})
	}
	return p
}

// teleports says where each teleport of the map leads: to the one place there is, or to one of several, chosen at random. The
// places where a teleport lands lead nowhere, and have no line.
func teleports(m domain.ExpeditionMap) []string {
	var lines []string
	for _, p := range m.Portals {
		switch n := len(p.To); n {
		case 0:
		case 1:
			lines = append(lines, p.Label+" takes you to "+p.To[0])
		default:
			lines = append(lines, p.Label+" takes you to "+strings.Join(p.To[:n-1], ", ")+" or "+p.To[n-1]+", at random")
		}
	}
	return lines
}

// chance says a chance in percent to one decimal, without the decimal when it is 0: 25%, 18.8%.
func chance(percent float64) string {
	return strconv.FormatFloat(math.Round(percent*10)/10, 'f', -1, 64) + "%"
}

// standing is what the game keeps of a monster on its trip, for boxes above the map: the energy that it has, coloured
// by how near it is to running out, where it is as the game shows it, and the steps that it has made.
func standing(r player.Ranch) []profileui.Box {
	var boxes []profileui.Box
	if r.EnergyKnown {
		tone := profileui.ToneLightGreen
		switch {
		case r.Energy < 20:
			tone = profileui.ToneOrangeRed
		case r.Energy < 100:
			tone = profileui.ToneOrange
		}
		boxes = append(boxes, profileui.Box{Text: fmt.Sprintf("Energy %d", r.Energy), Tone: tone})
	}
	x, y := r.Walked()
	steps := fmt.Sprintf("%d steps", r.Steps)
	if r.Steps == 1 {
		steps = "1 step"
	}
	return append(boxes,
		profileui.Box{Text: fmt.Sprintf("X %d · Y %d", x, y), Tone: profileui.ToneHotPink}, // the colour of its tile on the map
		profileui.Box{Text: steps, Tone: profileui.ToneLightYellow})
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
