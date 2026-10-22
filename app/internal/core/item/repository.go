package item

import (
	"context"

	"app/internal/core/item/domain"
)

type ItemRepository interface {
	ListItems(ctx context.Context) ([]domain.Item, error)
}
