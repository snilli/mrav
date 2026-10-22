package move

import (
	"context"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
)

type MoveRepository interface {
	ListGenera(ctx context.Context) ([]domain.Genus, error)
	GetGenus(ctx context.Context, code monster.Genus) (domain.Genus, error)
	ListMoves(ctx context.Context, code monster.Genus) ([]domain.Move, error)
}
