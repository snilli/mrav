package domain

import (
	"math"
	"slices"

	"app/internal/core/apperr"
	player "app/internal/core/player/domain"
)

// Terrain is what a tile of an expedition's map is to walk on.
type Terrain int

const (
	Floor    Terrain = iota
	Wall             // cannot be walked
	Rock             // an obstacle: each attempt to break it costs ObstacleCost, and can fail
	Gate             // a hidden way that a monster with enough intelligence can find: see ExpeditionMap.Locks
	Teleport         // leads to another tile of the map at random, so a route does not use it
)

// Point is a tile of a map, from the top left corner.
type Point struct{ X, Y int }

// Lock is what a Gate tile of a map wants: the intelligence that finding the way through it takes.
type Lock struct {
	At     Point
	MinINT int
}

// Spot is a place where the monster can search, and what a search there is worth to it on average, in gold. Loot is
// what a search there can turn up, from which Value is worked out for the monster that goes; it is nil where Value
// is given as it is.
type Spot struct {
	Name  string
	At    Point
	Value float64
	Loot  *Loot
}

// ExpeditionMap is the map of an expedition: the tiles row by row, where the monster starts, its places, and the
// locks that the gates have. Source says where the map is from, for the credit it is owed.
type ExpeditionMap struct {
	Name          string
	Source        string
	Width, Height int
	Tiles         []Terrain
	Start         Point
	Spots         []Spot
	Locks         []Lock
}

// Walker is a monster on an expedition, as far as a route goes: the energy it has, the chance that an attempt on a
// rock breaks it, and the intelligence that finds the way through a gate. At is the tile it stands on, or nil when it
// is at the start; the route ends at the start either way.
type Walker struct {
	Energy       int
	BreakChance  float64
	Intelligence int
	At           *Point
}

// Route is the places to search, in the order to go, and the walk that takes the monster to them and back to the
// start, which the expedition has to end on. Cost is the energy that it takes on average.
type Route struct {
	Spots []Spot
	Path  []Point
	Cost  float64
	Value float64
}

// maxSpots is how many places a map can have for PlanRoute: it weighs every set of them.
const maxSpots = 16

// PlanRoute finds the route that is worth most in what its places are worth, and for as little of the energy as
// that takes, among those that cost no more than the walker's energy.
//
// shortcut: a rock costs what it takes on average to break it every time it is crossed, because the game's
// rocks come back and when is not known; lower that if one stays broken for the trip. Likewise a gate costs what
// finding it takes on average each time, though the game keeps a way that has been found.
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
	// the point where the walk begins, then the places, then the start, which the trip has to end at
	points := append(append([]Point{origin}, spotPoints(m.Spots)...), m.Start)
	base := n + 1
	dist := make([][]float64, len(points)) // dist[i][j] is the cost from point i to point j, without a search
	prev := make([][]int, len(points))     // prev[i] is, for each tile, the tile it is reached from when leaving point i
	for i, p := range points {
		d, from := m.walk(p, steps)
		dist[i], prev[i] = make([]float64, len(points)), from
		for j, q := range points {
			dist[i][j] = d[m.index(q)]
		}
	}

	// cost[mask*n+last] is the least the walk costs from the start to search every place of mask, the last at last.
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
		for last := n - 1; last >= 0; last-- { // from the last place of the map, so that a tie keeps the map's order
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

	// The places in the order to go, read back from the last, and the walk between them.
	order := make([]int, 0, n)
	for mask, last := best.mask, best.last; ; {
		order = append(order, last)
		if mask&(mask-1) == 0 { // the first place has none before it
			break
		}
		mask, last = mask&^(1<<last), int(from[mask*n+last])
	}
	r := Route{Path: []Point{origin}, Cost: best.cost, Value: best.value}
	at := 0
	for _, o := range slices.Backward(order) {
		r.Spots = append(r.Spots, m.Spots[o])
		r.Path = append(r.Path, m.trail(prev[at], points[at], points[o+1])...)
		at = o + 1
	}
	r.Path = append(r.Path, m.trail(prev[at], points[at], points[base])...)
	return r, nil
}

func spotPoints(spots []Spot) []Point {
	out := make([]Point, len(spots))
	for i, s := range spots {
		out[i] = s.At
	}
	return out
}

// Here says whether the monster can be on the trip of this map: the week that the trip comes round has come, and the
// tile that the monster last stood on is a tile of the map that can be walked. The game keeps that tile after a trip
// as well, so this is the week that decides when it is not on one; where the week or the tile is not known, it is
// false.
//
// shortcut: whether the date holds still while a trip is made is not known; if it does not, the map goes away in
// the middle of the trip, and the tile of the monster would have to be what decides.
func (m ExpeditionMap) Here(r player.Ranch) bool {
	if _, ok := m.Stand(r); !ok {
		return false
	}
	now := r.Calendar()
	for _, e := range Expeditions {
		if e.Name == m.Name {
			return e.Month == now.Month && e.Week == now.Week
		}
	}
	return false
}

// Stand is the tile of the map that the monster last stood on, as far from the start of the map as it is from the
// start of the trip, and whether that is a tile of the map that can be walked. It is not where the place of the
// monster is not known.
func (m ExpeditionMap) Stand(r player.Ranch) (Point, bool) {
	if !r.Known || !r.WalkKnown {
		return Point{}, false
	}
	right, up := r.Walked()
	at := Point{X: m.Start.X + right, Y: m.Start.Y - up}
	if !m.inside(at) || m.Tiles[m.index(at)] == Wall {
		return Point{}, false
	}
	return at, true
}

func (m ExpeditionMap) index(p Point) int { return p.Y*m.Width + p.X }

func (m ExpeditionMap) inside(p Point) bool {
	return p.X >= 0 && p.X < m.Width && p.Y >= 0 && p.Y < m.Height
}

// check says whether the map is what it claims to be.
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

// stepCosts is what it costs the walker to step onto each tile: the step, and for a rock what it takes on average
// to break it, and for a gate what finding the way takes on average, or nothing at all where the walker cannot.
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

// walk is the cheapest way from a point to every tile of the map, by what it costs to step onto each: what it costs,
// and the tile each is reached from.
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

// trail is the tiles walked from one point to another, leaving out the one it starts on; from is what walk gave
// for the first.
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
