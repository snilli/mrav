// Package item is the Items page's use case: what the player carries, with what the game's list says of it.
package item

import (
	"context"

	"app/internal/core/item/domain"
)

// ItemRepository is the read port of the game's list of items.
type ItemRepository interface {
	// ListItems returns every item, ordered by code.
	ListItems(ctx context.Context) ([]domain.Item, error)
}
