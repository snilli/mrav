package move

import (
	"context"

	"app/internal/core/move/domain"
)

type MoveService interface {
	CurrentGrid(ctx context.Context) (domain.Grid, error)
	ListGenera(ctx context.Context) ([]domain.Genus, error)
}
