package domain

import (
	"math"
	"slices"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
)

type Terrain int

const (
	Floor Terrain = iota
	Wall
	Rock
	Gate
	Teleport
)

type Point struct{ X, Y int }

type Lock struct {
	At     Point
	MinINT int
}

type Spot struct {
	Name     string
	At       Point
	Building string
	BySkill  bool
	Chances  []Chance
	Value    float64
	Loot     []Drop
}

type Drop struct {
	Code  int
	Item  string
	Pct   float64
	Price int
}

type Searcher struct {
	Intelligence, Skill int
	GoodNature          bool
}

type ExpeditionMap struct {
	Name          string
	Width, Height int
	Tiles         []Terrain
	Start         Point
	Spots         []Spot
	Locks         []Lock
	Portals       []Portal
}

type Portal struct {
	Label string
	At    Point
	To    []string
}

type Walker struct {
	Energy       int
	BreakChance  float64
	Intelligence int
	At           *Point
}

type Route struct {
	Spots []Spot
	Path  []Point
	Cost  float64
	Value float64
}

const maxSpots = 16

func PlanRoute(m ExpeditionMap, w Walker) (Route, error) {
	if err := m.check(); err != nil {
		return Route{}, err
	}
	if w.At != nil && !m.inside(*w.At) {
		return Route{}, apperr.Invalid("walker", "is not on the map")
	}
	n := len(m.Spots)
	energy, steps := w.Energy, m.stepCosts(w)
	origin := m.Start
	if w.At != nil {
		origin = *w.At
	}
	points := []Point{origin}
	stands := [][]Point{{origin}}
	for _, s := range m.Spots {
		points, stands = append(points, s.At), append(stands, m.beside(s.At, steps))
	}
	points, stands = append(points, m.Start), append(stands, []Point{m.Start})
	base := n + 1
	dist := make([][]float64, len(stands))
	land := make([][]Point, len(stands))
	prev := make([][]int, len(stands))
	for i, p := range points {
		d, from := m.walk(p, steps)
		dist[i], land[i], prev[i] = make([]float64, len(stands)), make([]Point, len(stands)), from
		for j := range stands {
			dist[i][j], land[i][j] = m.nearest(d, stands[j])
		}
	}

	cost := make([]float64, (1<<n)*n)
	from := make([]int8, (1<<n)*n)
	for i := range cost {
		cost[i] = math.Inf(1)
	}
	for last := range n {
		cost[(1<<last)*n+last] = dist[0][last+1] + SearchCost
	}
	value := make([]float64, 1<<n)
	var best struct {
		mask, last  int
		value, cost float64
	}
	best.last = -1
	for mask := 1; mask < 1<<n; mask++ {
		for i := range n {
			if mask&(1<<i) != 0 {
				value[mask] += m.Spots[i].Value
			}
		}
		for last := n - 1; last >= 0; last-- {
			c := cost[mask*n+last]
			if mask&(1<<last) == 0 || math.IsInf(c, 1) {
				continue
			}
			for next := range n {
				if mask&(1<<next) != 0 {
					continue
				}
				nc := c + dist[last+1][next+1] + SearchCost
				if at := (mask|1<<next)*n + next; nc < cost[at] {
					cost[at], from[at] = nc, int8(last)
				}
			}
			total := c + dist[last+1][base]
			if total > float64(energy) || math.IsInf(total, 1) {
				continue
			}
			const eps = 1e-9
			if v := value[mask]; v > best.value+eps || (v >= best.value-eps && best.last >= 0 && total < best.cost-eps) {
				best.mask, best.last, best.value, best.cost = mask, last, v, total
			}
		}
	}
	if best.last < 0 {
		return Route{}, nil
	}

	order := make([]int, 0, n)
	for mask, last := best.mask, best.last; ; {
		order = append(order, last)
		if mask&(mask-1) == 0 {
			break
		}
		mask, last = mask&^(1<<last), int(from[mask*n+last])
	}
	r := Route{Path: []Point{origin}, Cost: best.cost, Value: best.value}
	at := 0
	for _, o := range slices.Backward(order) {
		r.Spots = append(r.Spots, m.Spots[o])
		r.Path = append(r.Path, m.trail(prev[at], points[at], land[at][o+1])...)
		at = o + 1
	}
	r.Path = append(r.Path, m.trail(prev[at], points[at], land[at][base])...)
	return r, nil
}

type Way struct {
	Steps int
	Rocks int
	Gates int
	Cost  float64
}

func (m ExpeditionMap) WaysTo(w Walker) map[string]Way {
	if m.check() != nil {
		return nil
	}
	origin := m.Start
	if w.At != nil && m.inside(*w.At) {
		origin = *w.At
	}
	steps := m.stepCosts(w)
	dist, from := m.walk(origin, steps)
	out := map[string]Way{}
	for _, s := range m.Spots {
		d, at := m.nearest(dist, m.beside(s.At, steps))
		if math.IsInf(d, 1) {
			continue
		}
		way := Way{Cost: d + SearchCost}
		for _, p := range m.trail(from, origin, at) {
			way.Steps++
			switch m.Tiles[m.index(p)] {
			case Rock:
				way.Rocks++
			case Gate:
				way.Gates++
			}
		}
		out[s.Name] = way
	}
	return out
}

func Near(a, b Point) bool {
	return max(a.X-b.X, b.X-a.X) <= 1 && max(a.Y-b.Y, b.Y-a.Y) <= 1
}

func (m ExpeditionMap) beside(p Point, steps []float64) []Point {
	var out []Point
	for y := p.Y - 1; y <= p.Y+1; y++ {
		for x := p.X - 1; x <= p.X+1; x++ {
			if q := (Point{X: x, Y: y}); m.inside(q) && !math.IsInf(steps[m.index(q)], 1) {
				out = append(out, q)
			}
		}
	}
	return out
}

func (m ExpeditionMap) nearest(dist []float64, at []Point) (float64, Point) {
	best, where := math.Inf(1), Point{}
	for _, p := range at {
		if d := dist[m.index(p)]; d < best {
			best, where = d, p
		}
	}
	return best, where
}

func (m ExpeditionMap) Here(r player.Ranch) bool {
	if _, ok := m.Stand(r); !ok {
		return false
	}
	e, ok := m.Trip()
	now := r.Calendar()
	return ok && e.Month == now.Month && e.Week == now.Week
}

func (m ExpeditionMap) Trip() (Expedition, bool) {
	if m.Name == Phoenix.Name {
		return Phoenix, true
	}
	for _, e := range Expeditions {
		if e.Name == m.Name {
			return e, true
		}
	}
	return Expedition{}, false
}

func (m ExpeditionMap) Open(r player.Ranch, mon monster.Monster) bool {
	e, ok := m.Trip()
	if !ok || !r.Known {
		return false
	}
	if (e.Name == Phoenix.Name) != PhoenixPending(r) {
		return false
	}
	now := r.Calendar()
	return e.Month == now.Month && e.Week == now.Week && e.Fit(mon)
}

func PhoenixPending(r player.Ranch) bool {
	if !r.PhoenixKnown {
		return false
	}
	now := r.Calendar()
	return !r.PhoenixDone || now.Month == Phoenix.Month && now.Week == Phoenix.Week && r.OnTrip(Phoenix.Origin.X, Phoenix.Origin.Y)
}

func (m ExpeditionMap) Stand(r player.Ranch) (Point, bool) {
	if !r.Known || !r.WalkKnown {
		return Point{}, false
	}
	right, up := m.Walked(r)
	at := Point{X: m.Start.X + right, Y: m.Start.Y - up}
	if !m.inside(at) || m.Tiles[m.index(at)] == Wall {
		return Point{}, false
	}
	return at, true
}

func (m ExpeditionMap) Walked(r player.Ranch) (right, up int) {
	if r.OffsetKnown {
		return r.OffsetX, r.OffsetY
	}
	e, ok := m.Trip()
	if !ok || e.Origin == (Point{}) {
		return r.Walked()
	}
	return r.Tile%player.MapWidth - e.Origin.X, e.Origin.Y - r.Tile/player.MapWidth
}

func (m ExpeditionMap) index(p Point) int { return p.Y*m.Width + p.X }

func (m ExpeditionMap) inside(p Point) bool {
	return p.X >= 0 && p.X < m.Width && p.Y >= 0 && p.Y < m.Height
}

func (m ExpeditionMap) check() error {
	switch {
	case m.Width <= 0 || m.Height <= 0 || len(m.Tiles) != m.Width*m.Height:
		return apperr.Invalid("tiles", "are not as many as the width and the height make")
	case !m.inside(m.Start):
		return apperr.Invalid("start", "is not on the map")
	case len(m.Spots) > maxSpots:
		return apperr.Invalid("spots", "are more than a route can weigh")
	}
	for _, s := range m.Spots {
		if !m.inside(s.At) {
			return apperr.Invalid("spots", s.Name+" is not on the map")
		}
	}
	locked := make(map[Point]bool, len(m.Locks))
	for _, l := range m.Locks {
		if !m.inside(l.At) || m.Tiles[m.index(l.At)] != Gate {
			return apperr.Invalid("locks", "has one that is not on a gate tile")
		}
		locked[l.At] = true
	}
	for i, t := range m.Tiles {
		if t == Gate && !locked[Point{X: i % m.Width, Y: i / m.Width}] {
			return apperr.Invalid("locks", "are missing for a gate tile")
		}
	}
	return nil
}

func (m ExpeditionMap) stepCosts(w Walker) []float64 {
	need := make(map[Point]int, len(m.Locks))
	for _, l := range m.Locks {
		need[l.At] = l.MinINT
	}
	costs := make([]float64, len(m.Tiles))
	for i, t := range m.Tiles {
		switch {
		case t == Wall, t == Teleport, t == Rock && w.BreakChance <= 0:
			costs[i] = math.Inf(1)
		case t == Rock:
			costs[i] = MoveCost + ObstacleCost/w.BreakChance
		case t == Gate && w.Intelligence < need[Point{X: i % m.Width, Y: i / m.Width}]:
			costs[i] = math.Inf(1)
		case t == Gate:
			costs[i] = MoveCost + SearchCost/FindChance(w.Intelligence)
		default:
			costs[i] = MoveCost
		}
	}
	return costs
}

func (m ExpeditionMap) walk(start Point, steps []float64) (dist []float64, from []int) {
	dist, from = make([]float64, len(m.Tiles)), make([]int, len(m.Tiles))
	for i := range dist {
		dist[i], from[i] = math.Inf(1), -1
	}
	dist[m.index(start)] = 0
	queue := []Point{start}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, step := range [...]Point{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			q := Point{p.X + step.X, p.Y + step.Y}
			if !m.inside(q) {
				continue
			}
			if c := dist[m.index(p)] + steps[m.index(q)]; c < dist[m.index(q)] {
				dist[m.index(q)], from[m.index(q)] = c, m.index(p)
				queue = append(queue, q)
			}
		}
	}
	return dist, from
}

func (m ExpeditionMap) trail(from []int, a, b Point) []Point {
	var rev []Point
	for t := m.index(b); t != m.index(a); t = from[t] {
		rev = append(rev, Point{X: t % m.Width, Y: t / m.Width})
	}
	out := make([]Point, len(rev))
	for i, p := range rev {
		out[len(rev)-1-i] = p
	}
	return out
}
