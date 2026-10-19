// Package fakeplayer is a PlayerSource whose player is set by hand, for the demo panel and for tests.
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

// New is a player with no money and an empty place for every item.
func New() *Source {
	var p player.Player
	for slot := range p.Items {
		p.Items[slot] = item.EmptyFrom
	}
	return &Source{p: p}
}

// Player is a copy: a Player holds only numbers, strings and arrays, so a reader cannot change the source's.
func (s *Source) Player(_ context.Context) (player.Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.p, nil
}

// Update changes the player the source shows.
func (s *Source) Update(change func(*player.Player)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	change(&s.p)
}
