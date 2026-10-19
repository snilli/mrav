// Package port holds the interfaces of the systems outside the app.
package port

import (
	"context"

	monster "app/internal/core/monster/domain"
)

// MonsterSource is where the monster being raised comes from: the game's memory, or a stand-in.
type MonsterSource interface {
	// Current is the monster the game shows now. It is apperr.ErrNotFound when the game shows none.
	Current(ctx context.Context) (monster.Monster, error)
}
