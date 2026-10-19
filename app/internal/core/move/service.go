package move

import (
	"context"

	"app/internal/core/move/domain"
)

// MoveService is the move viewer's use cases, one file each in service/movesvc.
type MoveService interface {
	// CurrentGrid is the grid of the monster the game shows now. It is apperr.ErrNotFound when the game
	// shows no monster, or one of a genus the catalog does not have.
	CurrentGrid(ctx context.Context) (domain.Grid, error)
	// ListGenera returns every genus, ordered by code.
	ListGenera(ctx context.Context) ([]domain.Genus, error)
}
