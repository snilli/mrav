package drill

import (
	"context"

	"app/internal/core/drill/domain"
)

type DrillService interface {
	Current(ctx context.Context) (domain.Training, error)
}
