package profilerepo

import (
	"context"
	"fmt"

	"orm"
	"orm/ent"
	entbreed "orm/ent/breed"
	entcocoon "orm/ent/cocoon"

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

// growthOf is the six growth rates a rule wants, in the order of monster.Stat.
func growthOf(rates []int) ([monster.StatCount]int, error) {
	var growth [monster.StatCount]int
	if len(rates) != monster.StatCount {
		return growth, fmt.Errorf("growth has %d rates, not %d", len(rates), monster.StatCount)
	}
	copy(growth[:], rates)
	return growth, nil
}
