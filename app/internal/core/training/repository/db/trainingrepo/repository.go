package trainingrepo

import (
	"context"

	"orm"
	"orm/ent"
	entgain "orm/ent/traininggain"

	"app/internal/core/training"
	"app/internal/core/training/domain"
	"app/internal/infrastructure/memo"
)

type Repository struct {
	client *orm.Client
	gains  memo.Cache[struct{}, []domain.Gain]
}

func New(client *orm.Client) training.TrainingRepository {
	return &Repository{client: client}
}

func (r *Repository) ListGains(ctx context.Context) ([]domain.Gain, error) {
	return r.gains.Get(struct{}{}, func() ([]domain.Gain, error) {
		rows, err := r.client.TrainingGain.Query().
			Order(ent.Asc(entgain.FieldLifeStage), ent.Asc(entgain.FieldGrowthRate), ent.Asc(entgain.FieldKind)).
			All(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]domain.Gain, len(rows))
		for i, g := range rows {
			out[i] = domain.Gain{LifeStage: g.LifeStage, GrowthRate: g.GrowthRate, Kind: g.Kind, Text: g.Text}
		}
		return out, nil
	})
}
