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

type View struct {
	Name         string
	Summary      string
	Places       []PlaceView
	Teleports    []string
	Intelligence int
	Marker       *domain.Point
	Standing     []profileui.Box
	Map          domain.ExpeditionMap
	Route        domain.Route
}

func NewView(p domain.Profile, m domain.ExpeditionMap) (View, error) {
	mon := p.Monster
	trip, _ := m.Trip()
	walker := domain.Walker{
		Energy:       domain.StartEnergy(mon.Stats[monster.Life]),
		BreakChance:  trip.BreakChance(mon.Stats[monster.Pow]),
		Intelligence: mon.Stats[monster.Int],
	}
	var marker *domain.Point
	left := false
	ranch := p.Player.Ranch
	if m.Here(ranch) {
		at, _ := m.Stand(ranch)
		marker = &at
		if left = ranch.EnergyKnown && ranch.Energy < walker.Energy; left {
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
		v.Standing = standing(ranch, m)
	}
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

type PlaceView struct {
	Name     string
	Building string
	Stop     int
	Worth    string
	Loot     []DropView
}

type DropView struct {
	Item   string
	Chance string
	Story  bool
}

var storyItems = map[string]bool{
	"Old Sheath": true, "Mask": true, "Big Footstep": true, "Fire Feather": true, "Strong Glue": true, "Gold Peach": true,
}

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

func chance(percent float64) string {
	return strconv.FormatFloat(math.Round(percent*10)/10, 'f', -1, 64) + "%"
}

func standing(r player.Ranch, m domain.ExpeditionMap) []profileui.Box {
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
	x, y := m.Walked(r)
	return append(boxes, profileui.Box{Text: fmt.Sprintf("X %d · Y %d", x, y), Tone: profileui.ToneHotPink})
}

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
