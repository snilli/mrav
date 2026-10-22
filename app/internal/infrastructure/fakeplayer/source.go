package fakeplayer

import (
	"context"
	"sync"

	item "app/internal/core/item/domain"
	player "app/internal/core/player/domain"
	"app/internal/port"
)

type Source struct {
	mu sync.RWMutex
	p  player.Player
}

var _ port.PlayerSource = (*Source)(nil)

func New() *Source {
	var p player.Player
	for slot := range p.Items {
		p.Items[slot] = item.EmptyFrom
	}
	return &Source{p: p}
}

func (s *Source) Player(_ context.Context) (player.Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.p, nil
}

func (s *Source) Update(change func(*player.Player)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	change(&s.p)
}
