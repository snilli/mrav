package drill

import (
	"context"

	"app/internal/core/drill/domain"
)

type DrillRepository interface {
	ListGains(ctx context.Context) ([]domain.Cell, error)
}
