package training

import (
	"context"

	"app/internal/core/training/domain"
)

type TrainingService interface {
	Current(ctx context.Context) (domain.Training, error)
}
