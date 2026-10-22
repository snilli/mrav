package port

import (
	"context"

	monster "app/internal/core/monster/domain"
)

type MonsterSource interface {
	Current(ctx context.Context) (monster.Monster, error)
}
