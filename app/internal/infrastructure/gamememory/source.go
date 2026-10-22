package gamememory

import (
	"context"
	"errors"
	"sync"
	"time"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/port"
)

type Memory interface {
	Read(addr uint32, p []byte) error
	Close() error
}

type ModuleMemory interface {
	Memory
	ReadModule(offset uint32, p []byte) error
}

type Attach func() (Memory, error)

type Source struct {
	layout Layout
	attach Attach
	now    func() time.Time

	mu       sync.Mutex
	mem      Memory
	failure  error
	failedAt time.Time
}

const retryAttach = time.Second

var (
	_ port.MonsterSource = (*Source)(nil)
	_ port.PlayerSource  = (*Source)(nil)
)

func New(layout Layout, attach Attach) *Source {
	return &Source{layout: layout, attach: attach, now: time.Now}
}

func DuckStation() *Source { return New(ps1, attachDuckStation) }

func MR2DX() *Source { return New(steam, attachMR2DX) }

func (s *Source) Current(_ context.Context) (monster.Monster, error) {
	var m monster.Monster
	err := s.read(func(mem Memory) (err error) {
		m, err = decode(s.layout, mem)
		return err
	})
	return m, err
}

func (s *Source) Player(_ context.Context) (player.Player, error) {
	var p player.Player
	err := s.read(func(mem Memory) (err error) {
		p, err = decodePlayer(s.layout, mem)
		return err
	})
	return p, err
}

func (s *Source) isAttached() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mem != nil
}

func (s *Source) read(fn func(Memory) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.mem == nil {
		if s.failure != nil && s.now().Sub(s.failedAt) < retryAttach {
			return s.failure
		}
		mem, err := s.attach()
		if err != nil {
			s.failure, s.failedAt = apperr.NotFound("game", err.Error()), s.now()
			return s.failure
		}
		s.failure, s.mem = nil, mem
	}
	err := fn(s.mem)
	if err != nil && !errors.Is(err, apperr.ErrNotFound) {
		_ = s.mem.Close()
		s.mem = nil
		return apperr.NotFound("game", "cannot read the game: "+err.Error())
	}
	return err
}
