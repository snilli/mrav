package training

import (
	"context"

	"app/internal/core/training/domain"
)

type TrainingRepository interface {
	ListGains(ctx context.Context) ([]domain.Gain, error)
}
