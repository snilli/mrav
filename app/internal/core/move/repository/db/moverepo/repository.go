package moverepo

import (
	"orm"

	"app/internal/core/move"
)

// Repository reads the move catalog through Ent's query builder only; the Ent-to-domain mapping is in
// mapper.go.
type Repository struct {
	client *orm.Client
}

func New(client *orm.Client) move.MoveRepository {
	return &Repository{client: client}
}
