package moverepo

import (
	"context"
	"fmt"

	"orm/ent"
	entgenus "orm/ent/genus"
	entmove "orm/ent/move"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
)

func (r *Repository) ListMoves(ctx context.Context, code monster.Genus) ([]domain.Move, error) {
	rows, err := r.client.Move.Query().
		Where(entmove.HasGenusWith(entgenus.Code(int(code)))).
		Order(ent.Asc(entmove.FieldMemorySlot)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Move, len(rows))
	for i, e := range rows {
		m, err := toMove(e, code)
		if err != nil {
			return nil, fmt.Errorf("move %q of genus %d: %w", e.Name, code, err)
		}
		out[i] = m
	}
	return out, nil
}
