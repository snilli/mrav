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
	return r.moves.Get(code, func() ([]domain.Move, error) { return r.loadMoves(ctx, code) })
}

// loadMoves finds the genus by its code and then its moves by the id of the genus: two lookups by an index, the second in the
// order of its memory slots. One statement over both tables would have the database read every move to see whose genus it is.
func (r *Repository) loadMoves(ctx context.Context, code monster.Genus) ([]domain.Move, error) {
	id, err := r.client.Genus.Query().Where(entgenus.Code(int(code))).OnlyID(ctx)
	if ent.IsNotFound(err) {
		return nil, nil // a genus that is not there has no moves
	}
	if err != nil {
		return nil, err
	}
	rows, err := r.client.Move.Query().
		Where(entmove.GenusID(id)).
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
