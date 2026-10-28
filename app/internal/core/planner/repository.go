package planner

import (
	"context"

	"app/internal/core/planner/domain"
)

type PlannerRepository interface {
	ListFoodReactions(ctx context.Context) ([]domain.FoodReaction, error)
	ListEffects(ctx context.Context) ([]domain.Effect, error)
}
