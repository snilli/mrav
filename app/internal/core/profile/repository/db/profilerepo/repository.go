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
	entitem "orm/ent/item"
	entlootchance "orm/ent/lootchance"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile"
	"app/internal/core/profile/domain"
	"app/internal/infrastructure/memo"
)

type Repository struct {
	client   *orm.Client
	genera   memo.Cache[struct{}, map[monster.Genus]string]
	breeds   memo.Cache[breedsKey, []domain.BreedRule]
	cocoons  memo.Cache[monster.Version, []domain.CocoonRule]
	desires  memo.Cache[monster.Version, map[int]string]
	maps     memo.Cache[struct{}, []domain.ExpeditionMap]
	calendar memo.Cache[monster.Version, []domain.CalendarEntry]
}

type breedsKey struct {
	main    int
	version monster.Version
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

func (r *Repository) ListBreedRules(ctx context.Context, main int, version monster.Version) ([]domain.BreedRule, error) {
	return r.breeds.Get(breedsKey{main, version}, func() ([]domain.BreedRule, error) { return r.loadBreedRules(ctx, main, version) })
}

func (r *Repository) loadBreedRules(ctx context.Context, main int, version monster.Version) ([]domain.BreedRule, error) {
	rows, err := r.client.Breed.Query().
		Where(entbreed.VersionEQ(entbreed.Version(version)), entbreed.MainGenus(main)).
		Order(ent.Asc(entbreed.FieldSubGenus), ent.Asc(entbreed.FieldSeq)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	rules := make([]domain.BreedRule, len(rows))
	for i, b := range rows {
		rule := domain.BreedRule{
			Main: b.MainGenus, Sub: b.SubGenus, Name: b.Name, Guts: b.Guts, MatchGuts: b.MatchGuts,
		}
		if len(b.Growth) > 0 {
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

func (r *Repository) ListCocoonRules(ctx context.Context, version monster.Version) ([]domain.CocoonRule, error) {
	return r.cocoons.Get(version, func() ([]domain.CocoonRule, error) { return r.loadCocoonRules(ctx, version) })
}

func (r *Repository) loadCocoonRules(ctx context.Context, version monster.Version) ([]domain.CocoonRule, error) {
	rows, err := r.client.Cocoon.Query().
		Where(entcocoon.VersionEQ(entcocoon.Version(version))).
		Order(ent.Asc(entcocoon.FieldID)).All(ctx)
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

func (r *Repository) ListDesireNames(ctx context.Context, version monster.Version) (map[int]string, error) {
	return r.desires.Get(version, func() (map[int]string, error) { return r.loadDesireNames(ctx, version) })
}

func (r *Repository) loadDesireNames(ctx context.Context, version monster.Version) (map[int]string, error) {
	rows, err := r.client.Item.Query().Where(entitem.VersionEQ(entitem.Version(version))).All(ctx)
	if err != nil {
		return nil, err
	}
	names := make(map[int]string, len(rows))
	for _, it := range rows {
		names[it.Code] = it.DesireName
	}
	return names, nil
}

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

	chances := make(map[int][]domain.Chance, len(places))
	for _, c := range rules {
		chances[c.PlaceID] = append(chances[c.PlaceID], domain.Chance{
			Good: c.Good, MinStat: c.MinStat, Slot: c.Slot, Code: c.Edges.Item.Code, Item: c.Edges.Item.Name, Pct: c.Pct, Price: price(c.Edges.Item.SalePrice),
		})
	}
	placesOf := make(map[int][]*ent.ExpeditionPlace, len(rows))
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

func price(sale *int) int {
	if sale == nil {
		return 0
	}
	return *sale
}

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
			default:
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

func growthOf(rates []int) ([monster.StatCount]int, error) {
	var growth [monster.StatCount]int
	if len(rates) != monster.StatCount {
		return growth, fmt.Errorf("growth has %d rates, not %d", len(rates), monster.StatCount)
	}
	copy(growth[:], rates)
	return growth, nil
}
