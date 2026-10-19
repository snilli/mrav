// Package fakemonster is a MonsterSource whose monster is set by hand, for the demo panel and for tests,
// until the game's memory can be read.
package fakemonster

import (
	"context"
	"sync"

	monster "app/internal/core/monster/domain"
	"app/internal/port"
)

type Source struct {
	mu sync.RWMutex
	m  monster.Monster
}

var _ port.MonsterSource = (*Source)(nil)

// New is a Pixie with no stats, no learned moves, and no active move.
func New() *Source {
	none := monster.NoActiveMove
	return &Source{m: monster.Monster{Active: [monster.ActiveGroups]int{none, none, none, none}}}
}

// Current is a copy: a Monster holds only numbers and arrays, so a reader cannot change the source's.
func (s *Source) Current(_ context.Context) (monster.Monster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.m, nil
}

// Update changes the monster the source shows.
func (s *Source) Update(change func(*monster.Monster)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	change(&s.m)
}
