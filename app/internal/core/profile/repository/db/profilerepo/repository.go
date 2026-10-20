package profilerepo

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"orm"
	"orm/ent"
	entbreed "orm/ent/breed"
	entcocoon "orm/ent/cocoon"
	entexpeditionlock "orm/ent/expeditionlock"
	entexpeditionmap "orm/ent/expeditionmap"
	entexpeditionplace "orm/ent/expeditionplace"
	entlootchance "orm/ent/lootchance"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile"
	"app/internal/core/profile/domain"
	"app/internal/infrastructure/memo"
)

// Repository reads the data the Monster page names things from through Ent's query builder. None of it changes while the
// viewer runs, so each answer is read once and kept: the pages ask twice a second, and what they are given is shared, so they
// must not write to it.
type Repository struct {
	client  *orm.Client
	genera  memo.Cache[struct{}, map[monster.Genus]string]
	breeds  memo.Cache[int, []domain.BreedRule]
	cocoons memo.Cache[struct{}, []domain.CocoonRule]
	desires memo.Cache[struct{}, map[int]string]
	maps    memo.Cache[struct{}, []domain.ExpeditionMap]
}

func New(client *orm.Client) profile.ProfileRepository {
	return &Repository{client: client}
}

func (r *Repository) ListGenusNames(ctx context.Context) (map[monster.Genus]string, error) {
	return r.genera.Get(struct{}{}, func() (map[monster.Genus]string, error) { return r.loadGenusNames(ctx) })
}

func (r *Repository) loadGenusNames(ctx context.Context) (map[monster.Genus]string, error) {
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
	return r.breeds.Get(main, func() ([]domain.BreedRule, error) { return r.loadBreedRules(ctx, main) })
}

func (r *Repository) loadBreedRules(ctx context.Context, main int) ([]domain.BreedRule, error) {
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
	return r.cocoons.Get(struct{}{}, func() ([]domain.CocoonRule, error) { return r.loadCocoonRules(ctx) })
}

func (r *Repository) loadCocoonRules(ctx context.Context) ([]domain.CocoonRule, error) {
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
	return r.desires.Get(struct{}{}, func() (map[int]string, error) { return r.loadDesireNames(ctx) })
}

func (r *Repository) loadDesireNames(ctx context.Context) (map[int]string, error) {
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

// ListExpeditionMaps reads the maps, their places, their locks and the tables of chances of the places, each with a
// statement that looks the rows up by an index and gives them in the order of it, and puts them together here: what a
// search is worth to a monster is the domain's to work out from the tables, not the database's.
func (r *Repository) ListExpeditionMaps(ctx context.Context) ([]domain.ExpeditionMap, error) {
	return r.maps.Get(struct{}{}, func() ([]domain.ExpeditionMap, error) { return r.loadExpeditionMaps(ctx) })
}

func (r *Repository) loadExpeditionMaps(ctx context.Context) ([]domain.ExpeditionMap, error) {
	rows, err := r.client.ExpeditionMap.Query().Order(ent.Asc(entexpeditionmap.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	places, err := r.client.ExpeditionPlace.Query().
		Order(ent.Asc(entexpeditionplace.FieldMapID), ent.Asc(entexpeditionplace.FieldName)).All(ctx)
	if err != nil {
		return nil, err
	}
	locks, err := r.client.ExpeditionLock.Query().
		Order(ent.Asc(entexpeditionlock.FieldMapID), ent.Asc(entexpeditionlock.FieldX), ent.Asc(entexpeditionlock.FieldY)).All(ctx)
	if err != nil {
		return nil, err
	}
	rules, err := r.client.LootChance.Query().WithItem().
		Order(ent.Asc(entlootchance.FieldPlaceID), ent.Asc(entlootchance.FieldGood), ent.Asc(entlootchance.FieldMinStat)).All(ctx)
	if err != nil {
		return nil, err
	}

	chances := make(map[int][]domain.Chance, len(places)) // by the id of the place
	for _, c := range rules {
		chances[c.PlaceID] = append(chances[c.PlaceID], domain.Chance{
			Good: c.Good, MinStat: c.MinStat, Slot: c.Slot, Item: c.Edges.Item.Name, Pct: c.Pct, Price: price(c.Edges.Item.SalePrice),
		})
	}
	placesOf := make(map[int][]*ent.ExpeditionPlace, len(rows)) // by the id of the map
	for _, p := range places {
		placesOf[p.MapID] = append(placesOf[p.MapID], p)
	}
	locksOf := make(map[int][]*ent.ExpeditionLock, len(rows))
	for _, l := range locks {
		locksOf[l.MapID] = append(locksOf[l.MapID], l)
	}

	maps := make([]domain.ExpeditionMap, len(rows))
	for i, row := range rows {
		m, err := mapOf(row, placesOf[row.ID], locksOf[row.ID], chances)
		if err != nil {
			return nil, fmt.Errorf("map %q: %w", row.Name, err)
		}
		maps[i] = m
	}
	return maps, nil
}

// price is what an item sells for, in gold: the price that the data has for it, 0 where it is not a number, as it is not for
// what is not sold.
func price(sale string) int {
	n, err := strconv.Atoi(sale)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

// mapOf reads the tiles of a stored map, and gives it its places with their tables of chances and its locks.
func mapOf(row *ent.ExpeditionMap, places []*ent.ExpeditionPlace, locks []*ent.ExpeditionLock, chances map[int][]domain.Chance) (domain.ExpeditionMap, error) {
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
	for _, l := range locks {
		m.Locks = append(m.Locks, domain.Lock{At: domain.Point{X: l.X, Y: l.Y}, MinINT: l.MinInt})
	}
	for _, p := range places {
		m.Spots = append(m.Spots, domain.Spot{
			Name: p.Name, At: domain.Point{X: p.X, Y: p.Y}, Building: p.Building,
			BySkill: p.Stat == entexpeditionplace.StatSkill, Chances: chances[p.ID],
		})
	}
	for _, p := range row.Portals {
		m.Portals = append(m.Portals, domain.Portal{Label: p.Label, At: domain.Point{X: p.X, Y: p.Y}, To: p.To})
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
