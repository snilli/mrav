package item

import (
	"context"

	"app/internal/core/item/domain"
	monster "app/internal/core/monster/domain"
)

type ItemRepository interface {
	ListItems(ctx context.Context, version monster.Version) ([]domain.Item, error)
}
