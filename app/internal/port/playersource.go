package port

import (
	"context"

	player "app/internal/core/player/domain"
)

// PlayerSource is where the player's side of the game comes from: the game's memory, or a stand-in.
type PlayerSource interface {
	// Player is the player's money, items and so on now. It is apperr.ErrNotFound when the game cannot be reached.
	Player(ctx context.Context) (player.Player, error)
}
