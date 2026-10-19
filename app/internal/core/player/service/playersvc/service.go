package playersvc

import (
	"context"

	"app/internal/core/player"
	"app/internal/core/player/domain"
	"app/internal/port"
)

type Service struct {
	source port.PlayerSource
}

func New(source port.PlayerSource) player.PlayerService {
	return &Service{source: source}
}

func (s *Service) Current(ctx context.Context) (domain.Player, error) {
	return s.source.Player(ctx)
}
