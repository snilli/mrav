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

func New() *Source {
	none := monster.NoActiveMove
	return &Source{m: monster.Monster{Active: [monster.ActiveGroups]int{none, none, none, none}}}
}

func (s *Source) Current(_ context.Context) (monster.Monster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.m, nil
}

func (s *Source) Update(change func(*monster.Monster)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	change(&s.m)
}
