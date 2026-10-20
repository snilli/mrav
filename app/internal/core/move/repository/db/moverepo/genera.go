package moverepo

import (
	"context"
	"fmt"

	"orm/ent"
	entgenus "orm/ent/genus"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
)

func (r *Repository) ListGenera(ctx context.Context) ([]domain.Genus, error) {
	return r.genera.Get(struct{}{}, func() ([]domain.Genus, error) {
		rows, err := r.client.Genus.Query().Order(ent.Asc(entgenus.FieldCode)).All(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]domain.Genus, len(rows))
		for i, g := range rows {
			out[i] = domain.Genus{Code: monster.Genus(g.Code), Name: g.Name}
		}
		return out, nil
	})
}

func (r *Repository) GetGenus(ctx context.Context, code monster.Genus) (domain.Genus, error) {
	return r.genus.Get(code, func() (domain.Genus, error) {
		g, err := r.client.Genus.Query().Where(entgenus.Code(int(code))).Only(ctx)
		if ent.IsNotFound(err) {
			return domain.Genus{}, apperr.NotFound("genus", fmt.Sprintf("no genus has code %d", code))
		}
		if err != nil {
			return domain.Genus{}, err
		}
		return domain.Genus{Code: monster.Genus(g.Code), Name: g.Name}, nil
	})
}
