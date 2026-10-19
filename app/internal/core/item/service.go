package item

import (
	"context"

	"app/internal/core/item/domain"
)

// ItemService is the Items page's use cases, in service/itemsvc.
type ItemService interface {
	// Inventory is what the player carries now, a slot for each of the places. It is apperr.ErrNotFound when the
	// game cannot be reached.
	Inventory(ctx context.Context) ([]domain.Slot, error)
}
