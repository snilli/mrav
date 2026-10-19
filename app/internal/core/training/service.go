package training

import (
	"context"

	"app/internal/core/training/domain"
)

// TrainingService is the Training page's use cases, in service/trainingsvc.
type TrainingService interface {
	// Current is the Training page for the monster the game shows now. It is apperr.ErrNotFound when the game
	// shows no monster.
	Current(ctx context.Context) (domain.Training, error)
}
