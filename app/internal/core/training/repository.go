// Package training is the Training page's use case: what each drill gives the monster the game shows.
package training

import (
	"context"

	"app/internal/core/training/domain"
)

// TrainingRepository is the read port of the table of training gains.
type TrainingRepository interface {
	// ListGains returns every gain, ordered by life stage, growth rate and kind.
	ListGains(ctx context.Context) ([]domain.Gain, error)
}
