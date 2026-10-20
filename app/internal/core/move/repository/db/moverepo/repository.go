package moverepo

import (
	"orm"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move"
	"app/internal/core/move/domain"
	"app/internal/infrastructure/memo"
)

// Repository reads the move catalog through Ent's query builder only; the Ent-to-domain mapping is in
// mapper.go. The catalog never changes while the viewer runs, so each answer is read once and kept: the pages ask for
// them twice a second, and what they are given is shared, so they must not write to it.
type Repository struct {
	client *orm.Client
	genera memo.Cache[struct{}, []domain.Genus]
	genus  memo.Cache[monster.Genus, domain.Genus]
	moves  memo.Cache[monster.Genus, []domain.Move]
}

func New(client *orm.Client) move.MoveRepository {
	return &Repository{client: client}
}
