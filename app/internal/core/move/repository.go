// Package move is the read side of the move catalog: a genus's moves, laid out for a monster.
package move

import (
	"context"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
)

// MoveRepository is the read port of the move catalog.
type MoveRepository interface {
	// ListGenera returns every genus, ordered by code.
	ListGenera(ctx context.Context) ([]domain.Genus, error)
	// GetGenus is apperr.ErrNotFound when no genus has the code.
	GetGenus(ctx context.Context, code monster.Genus) (domain.Genus, error)
	// ListMoves returns a genus's moves ordered by memory slot; a genus with no moves gives none.
	ListMoves(ctx context.Context, code monster.Genus) ([]domain.Move, error)
}
