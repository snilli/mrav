package itemsvc

import (
	"cmp"
	"context"

	"app/internal/core/item"
	"app/internal/core/item/domain"
	monster "app/internal/core/monster/domain"
	"app/internal/port"
)

type Service struct {
	repo    item.ItemRepository
	players port.PlayerSource
}

func New(repo item.ItemRepository, players port.PlayerSource) item.ItemService {
	return &Service{repo: repo, players: players}
}

func (s *Service) Inventory(ctx context.Context) ([]domain.Slot, error) {
	p, err := s.players.Player(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListItems(ctx, cmp.Or(p.Version, monster.MR2DX))
	if err != nil {
		return nil, err
	}
	return domain.Lock(domain.Fill(p.Items[:], items), p.Ranch.Slots()), nil
}
