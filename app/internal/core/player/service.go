// Package player is the use case of the pages that show the player's side of the game: the breeder and the controller.
package player

import (
	"context"

	"app/internal/core/player/domain"
)

// PlayerService is those pages' use cases, in service/playersvc.
type PlayerService interface {
	// Current is the player's side of the game now. It is apperr.ErrNotFound when the game cannot be reached.
	Current(ctx context.Context) (domain.Player, error)
}
