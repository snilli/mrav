package plannerrepo

import (
	"context"
	"fmt"

	"orm"
	"orm/ent"
	enteffect "orm/ent/effect"
	entfood "orm/ent/foodreaction"

	monster "app/internal/core/monster/domain"
	"app/internal/core/planner"
	"app/internal/core/planner/domain"
	"app/internal/infrastructure/memo"
)

type Repository struct {
	client    *orm.Client
	reactions memo.Cache[struct{}, []domain.FoodReaction]
	effects   memo.Cache[struct{}, []domain.Effect]
}

func New(client *orm.Client) planner.PlannerRepository {
	return &Repository{client: client}
}

func (r *Repository) ListFoodReactions(ctx context.Context) ([]domain.FoodReaction, error) {
	return r.reactions.Get(struct{}{}, func() ([]domain.FoodReaction, error) {
		rows, err := r.client.FoodReaction.Query().
			Order(ent.Asc(entfood.FieldGenus), ent.Asc(entfood.FieldFood), ent.Asc(entfood.FieldStage)).
			All(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]domain.FoodReaction, len(rows))
		for i, f := range rows {
			food, okFood := domain.ParseFood(string(f.Food))
			stage, okStage := domain.ParseStage(string(f.Stage))
			reaction, okReaction := domain.ParseReaction(string(f.Reaction))
			if !okFood || !okStage || !okReaction {
				return nil, fmt.Errorf("food reaction %d does not name a food, an age and a reaction: %q %q %q", f.ID, f.Food, f.Stage, f.Reaction)
			}
			out[i] = domain.FoodReaction{Genus: f.Genus, Food: food, Stage: stage, Reaction: reaction}
		}
		return out, nil
	})
}

func (r *Repository) ListEffects(ctx context.Context) ([]domain.Effect, error) {
	return r.effects.Get(struct{}{}, func() ([]domain.Effect, error) {
		rows, err := r.client.Effect.Query().
			Order(ent.Asc(enteffect.FieldSubject), ent.Asc(enteffect.FieldVersion), ent.Asc(enteffect.FieldGenus), ent.Asc(enteffect.FieldLifeStage)).
			All(ctx)
		if err != nil {
			return nil, err
		}
		out := make([]domain.Effect, len(rows))
		type key struct {
			subject, version string
			genus, stage     int
		}
		seen := make(map[key]bool, len(rows))
		for i, f := range rows {
			k := key{f.Subject, string(f.Version), valueOr(f.Genus, -1), valueOr(f.LifeStage, -1)}
			if !domain.KnownSubject(f.Subject) {
				return nil, fmt.Errorf("effect %d is for %q, which the planner does not know", f.ID, f.Subject)
			}
			if seen[k] {
				return nil, fmt.Errorf("effect %d is a second row for %q in %s with the same genus and life stage", f.ID, f.Subject, f.Version)
			}
			seen[k] = true
			out[i] = domain.Effect{
				Subject: f.Subject, Version: monster.Version(f.Version), Genus: f.Genus, LifeStage: f.LifeStage, StressFactor: f.StressFactor,
				Fatigue: f.Fatigue, Stress: f.Stress, FormDivisor: f.FormDivisor, Lost: f.Lost, Form: f.Form,
			}
		}
		return out, nil
	})
}

func valueOr(p *int, none int) int {
	if p == nil {
		return none
	}
	return *p
}
