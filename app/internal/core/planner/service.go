package planner

import (
	"context"

	"app/internal/core/planner/domain"
)

type PlannerService interface {
	Current(ctx context.Context) (domain.Current, error)
}
