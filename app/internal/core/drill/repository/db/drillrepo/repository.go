package drillrepo

import (
	"context"

	"orm"
	"orm/ent"
	entgain "orm/ent/drillgain"

	"app/internal/core/drill"
	"app/internal/core/drill/domain"
	"app/internal/infrastructure/memo"
)

type Repository struct {
	client *orm.Client
	cells  memo.Cache[struct{}, []domain.Cell]
}

func New(client *orm.Client) drill.DrillRepository {
	return &Repository{client: client}
}

func (r *Repository) ListGains(ctx context.Context) ([]domain.Cell, error) {
	return r.cells.Get(struct{}{}, func() ([]domain.Cell, error) {
		rows, err := r.client.DrillGain.Query().
			Order(
				ent.Asc(entgain.FieldResult), ent.Asc(entgain.FieldRole), ent.Asc(entgain.FieldBonus),
				ent.Asc(entgain.FieldLifeStage), ent.Asc(entgain.FieldGrowthRate),
			).
			All(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]domain.Cell, len(rows))
		for i, g := range rows {
			out[i] = domain.Cell{
				Result: domain.Result(g.Result), Role: domain.Role(g.Role), Bonus: g.Bonus,
				LifeStage: g.LifeStage, GrowthRate: g.GrowthRate, Slots: g.Slots,
			}
		}
		return out, nil
	})
}
