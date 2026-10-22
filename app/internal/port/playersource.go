package port

import (
	"context"

	player "app/internal/core/player/domain"
)

type PlayerSource interface {
	Player(ctx context.Context) (player.Player, error)
}
