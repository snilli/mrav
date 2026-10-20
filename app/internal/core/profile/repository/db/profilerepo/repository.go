package profilerepo

import (
	"context"
	"fmt"
	"strings"

	"orm"
	"orm/ent"
	entbreed "orm/ent/breed"
	entcocoon "orm/ent/cocoon"
	entexpeditionlock "orm/ent/expeditionlock"
	entexpeditionmap "orm/ent/expeditionmap"
	entexpeditionplace "orm/ent/expeditionplace"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile"
	"app/internal/core/profile/domain"
)

// Repository reads the data the Monster page names things from through Ent's query builder.
type Repository struct {
	client *orm.Client
}

func New(client *orm.Client) profile.ProfileRepository {
	return &Repository{client: client}
}

func (r *Repository) ListGenusNames(ctx context.Context) (map[monster.Genus]string, error) {
	rows, err := r.client.Genus.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	names := make(map[monster.Genus]string, len(rows))
	for _, g := range rows {
		names[monster.Genus(g.Code)] = g.Name
	}
	return names, nil
}

func (r *Repository) ListBreedRules(ctx context.Context, main int) ([]domain.BreedRule, error) {
	rows, err := r.client.Breed.Query().
		Where(entbreed.MainGenus(main)).
		Order(ent.Asc(entbreed.FieldSubGenus), ent.Asc(entbreed.FieldSeq)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	rules := make([]domain.BreedRule, len(rows))
	for i, b := range rows {
		rule := domain.BreedRule{
			Main: b.MainGenus, Sub: b.SubGenus, Name: b.Name, Otherwise: b.Otherwise, Guts: b.Guts, MatchGuts: b.MatchGuts,
		}
		if len(b.Growth) > 0 { // none means any growth rates will do
			growth, err := growthOf(b.Growth)
			if err != nil {
				return nil, fmt.Errorf("breed %d/%d: %w", b.MainGenus, b.SubGenus, err)
			}
			rule.Growth = &growth
		}
		rules[i] = rule
	}
	return rules, nil
}

func (r *Repository) ListCocoonRules(ctx context.Context) ([]domain.CocoonRule, error) {
	rows, err := r.client.Cocoon.Query().Order(ent.Asc(entcocoon.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	rules := make([]domain.CocoonRule, len(rows))
	for i, c := range rows {
		growth, err := growthOf(c.Growth)
		if err != nil {
			return nil, fmt.Errorf("cocoon %q: %w", c.Name, err)
		}
		rules[i] = domain.CocoonRule{Guts: c.Guts, Growth: growth, Name: c.Name}
	}
	return rules, nil
}

func (r *Repository) ListDesireNames(ctx context.Context) (map[int]string, error) {
	rows, err := r.client.Item.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	names := make(map[int]string, len(rows))
	for _, it := range rows {
		names[it.Code] = it.DesireName
	}
	return names, nil
}

// worthSQL works out what a search of each place is worth in gold on average, for a monster that searches as the
// arguments say: the chance of each item at the tier of the stat that the building is searched by, which is the
// highest tier that starts no higher than the stat, by what the item sells for, which is nothing where it is not
// sold. The items are joined by id, so that two items of a name are not counted for one chance. The arguments are
// whether the nature is good, the skill and the intelligence.
const worthSQL = `
SELECT p.id, SUM(c.pct / 100.0 * CASE WHEN i.sale_price GLOB '[0-9]*' THEN CAST(i.sale_price AS INTEGER) ELSE 0 END)
FROM expedition_places p
JOIN loot_chances c ON c.place_id = p.id AND c.good = ?
JOIN items i ON i.id = c.item_id
WHERE c.min_stat = (
	SELECT MAX(t.min_stat) FROM loot_chances t
	WHERE t.place_id = p.id AND t.good = c.good
	  AND t.min_stat <= CASE p.stat WHEN 'skill' THEN ? ELSE ? END)
GROUP BY p.id`

func (r *Repository) ListExpeditionMaps(ctx context.Context, by domain.Searcher) ([]domain.ExpeditionMap, error) {
	rows, err := r.client.ExpeditionMap.Query().
		Order(ent.Asc(entexpeditionmap.FieldID)).
		WithPlaces(func(q *ent.ExpeditionPlaceQuery) { q.Order(ent.Asc(entexpeditionplace.FieldName)) }).
		WithLocks(func(q *ent.ExpeditionLockQuery) { q.Order(ent.Asc(entexpeditionlock.FieldID)) }).
		All(ctx)
	if err != nil {
		return nil, err
	}
	worth, err := r.worths(ctx, by)
	if err != nil {
		return nil, err
	}
	maps := make([]domain.ExpeditionMap, len(rows))
	for i, row := range rows {
		m, err := mapOf(row, worth)
		if err != nil {
			return nil, fmt.Errorf("map %q: %w", row.Name, err)
		}
		maps[i] = m
	}
	return maps, nil
}

// worths is what a search of each place is worth, by the id of the place; a place without chances is not in it.
func (r *Repository) worths(ctx context.Context, by domain.Searcher) (map[int]float64, error) {
	rows, err := r.client.DB.QueryContext(ctx, worthSQL, by.GoodNature, by.Skill, by.Intelligence)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	worth := map[int]float64{}
	for rows.Next() {
		var id int
		var gold float64
		if err := rows.Scan(&id, &gold); err != nil {
			return nil, err
		}
		worth[id] = gold
	}
	return worth, rows.Err()
}

// mapOf reads the tiles of a stored map, and gives its places what a search of them is worth.
func mapOf(row *ent.ExpeditionMap, worth map[int]float64) (domain.ExpeditionMap, error) {
	lines := strings.Split(row.Rows, "\n")
	m := domain.ExpeditionMap{
		Name: row.Name, Width: len(lines[0]), Height: len(lines),
		Start: domain.Point{X: row.StartX, Y: row.StartY},
	}
	for y, line := range lines {
		if len(line) != m.Width {
			return domain.ExpeditionMap{}, fmt.Errorf("row %d is not as long as the first", y)
		}
		for _, c := range line {
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
	for _, l := range row.Edges.Locks {
		m.Locks = append(m.Locks, domain.Lock{At: domain.Point{X: l.X, Y: l.Y}, MinINT: l.MinInt})
	}
	for _, p := range row.Edges.Places {
		m.Spots = append(m.Spots, domain.Spot{
			Name: p.Name, At: domain.Point{X: p.X, Y: p.Y}, Building: p.Building, Value: worth[p.ID],
		})
	}
	return m, nil
}

// growthOf is the six growth rates a rule wants, in the order of monster.Stat.
func growthOf(rates []int) ([monster.StatCount]int, error) {
	var growth [monster.StatCount]int
	if len(rates) != monster.StatCount {
		return growth, fmt.Errorf("growth has %d rates, not %d", len(rates), monster.StatCount)
	}
	copy(growth[:], rates)
	return growth, nil
}
