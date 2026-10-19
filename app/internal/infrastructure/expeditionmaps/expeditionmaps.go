// Package expeditionmaps holds the maps of the expeditions that come round every year, Parepare, Torles and Kawrea,
// with the places on them and what a search of each can turn up. They are embedded in the viewer as the guide of
// LegendCup has them, and carry the credit that they are owed.
package expeditionmaps

import (
	_ "embed"
	"encoding/json"
	"fmt"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
)

//go:embed maps.json
var data []byte

// file is the shape of maps.json. A map is rows of tiles: # a wall, . a floor, S the start, o a rock, g a gate that
// wants intelligence, T a teleport, and a letter a place.
type file struct {
	Source string         `json:"source"`
	Prices map[string]int `json:"prices"` // what each item sells for
	Maps   []fileMap      `json:"maps"`
}

type fileMap struct {
	Name   string      `json:"name"`
	Rows   []string    `json:"rows"`
	Start  [2]int      `json:"start"`
	Gates  []fileGate  `json:"gates"`
	Places []filePlace `json:"places"`
}

type fileGate struct {
	At  [2]int `json:"at"`
	INT int    `json:"int"`
}

type filePlace struct {
	Name     string       `json:"name"`
	At       [2]int       `json:"at"`
	Building string       `json:"building"`
	Stat     string       `json:"stat"` // Intelligence or Skill
	Mins     []int        `json:"mins"`
	Bad      []fileChance `json:"bad"`
	Good     []fileChance `json:"good"`
}

type fileChance struct {
	Item string    `json:"item"`
	Pct  []float64 `json:"pct"`
}

// Load reads the maps that are embedded.
func Load() ([]domain.ExpeditionMap, error) {
	var f file
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("failed to read the expedition maps: %w", err)
	}
	maps := make([]domain.ExpeditionMap, 0, len(f.Maps))
	for _, fm := range f.Maps {
		m, err := build(f, fm)
		if err != nil {
			return nil, fmt.Errorf("failed to read the map of %s: %w", fm.Name, err)
		}
		maps = append(maps, m)
	}
	return maps, nil
}

func build(f file, fm fileMap) (domain.ExpeditionMap, error) {
	if len(fm.Rows) == 0 {
		return domain.ExpeditionMap{}, fmt.Errorf("it has no rows")
	}
	m := domain.ExpeditionMap{
		Name: fm.Name, Source: f.Source, Width: len(fm.Rows[0]), Height: len(fm.Rows),
		Start: domain.Point{X: fm.Start[0], Y: fm.Start[1]},
	}
	for y, row := range fm.Rows {
		if len(row) != m.Width {
			return domain.ExpeditionMap{}, fmt.Errorf("row %d is not as long as the first", y)
		}
		for _, c := range row {
			switch c {
			case '#':
				m.Tiles = append(m.Tiles, domain.Wall)
			case 'o':
				m.Tiles = append(m.Tiles, domain.Rock)
			case 'g':
				m.Tiles = append(m.Tiles, domain.Gate)
			case 'T':
				m.Tiles = append(m.Tiles, domain.Teleport)
			default: // a floor, the start, or a place
				m.Tiles = append(m.Tiles, domain.Floor)
			}
		}
	}
	for _, g := range fm.Gates {
		m.Locks = append(m.Locks, domain.Lock{At: domain.Point{X: g.At[0], Y: g.At[1]}, MinINT: g.INT})
	}
	for _, p := range fm.Places {
		at := domain.Point{X: p.At[0], Y: p.At[1]}
		if at.Y >= len(fm.Rows) || at.X >= m.Width || p.Name == "" || fm.Rows[at.Y][at.X] != p.Name[0] {
			return domain.ExpeditionMap{}, fmt.Errorf("place %s is not at %d,%d", p.Name, at.X, at.Y)
		}
		loot := domain.Loot{Building: p.Building, Stat: monster.Int, Mins: p.Mins, Bad: chances(f, p.Bad), Good: chances(f, p.Good)}
		if p.Stat == "Skill" {
			loot.Stat = monster.Skill
		}
		m.Spots = append(m.Spots, domain.Spot{Name: p.Name, At: at, Loot: &loot})
	}
	return m, nil
}

func chances(f file, in []fileChance) []domain.Chance {
	out := make([]domain.Chance, len(in))
	for i, c := range in {
		out[i] = domain.Chance{Item: c.Item, Price: f.Prices[c.Item], Pct: c.Pct}
	}
	return out
}
