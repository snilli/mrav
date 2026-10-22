package moverepo

import (
	"orm"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move"
	"app/internal/core/move/domain"
	"app/internal/infrastructure/memo"
)

type Repository struct {
	client *orm.Client
	genera memo.Cache[struct{}, []domain.Genus]
	genus  memo.Cache[monster.Genus, domain.Genus]
	moves  memo.Cache[monster.Genus, []domain.Move]
}

func New(client *orm.Client) move.MoveRepository {
	return &Repository{client: client}
}
