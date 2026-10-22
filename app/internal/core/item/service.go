package item

import (
	"context"

	"app/internal/core/item/domain"
)

type ItemService interface {
	Inventory(ctx context.Context) ([]domain.Slot, error)
}
