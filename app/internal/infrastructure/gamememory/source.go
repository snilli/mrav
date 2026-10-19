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

// Memory is the PlayStation's RAM inside a running game, addressed the way the console addresses it.
type Memory interface {
	Read(addr uint32, p []byte) error
	Close() error
}

// ModuleMemory is a Memory that can also read the game's own exe, outside the console's RAM, which is where the
// English release of MR2DX keeps what it adds to the console. Offsets are from where the exe is loaded.
type ModuleMemory interface {
	Memory
	ReadModule(offset uint32, p []byte) error
}

// Attach finds the running game and opens its RAM, or says why it cannot.
type Attach func() (Memory, error)

// Source reads a game's RAM. It attaches when it first needs to, and again whenever it has lost the game, so
// the game can be started, closed and started again under a viewer that stays open.
type Source struct {
	layout Layout
	attach Attach
	now    func() time.Time

	mu       sync.Mutex
	mem      Memory // nil while not attached
	failure  error  // why the last attach failed, to answer with until the next try
	failedAt time.Time
}

// retryAttach is how long a failed attach is answered from memory. Every page asks, several times a second, and
// an attach looks through the running programs, so with the game closed it would be most of the viewer's work.
const retryAttach = time.Second

var (
	_ port.MonsterSource = (*Source)(nil)
	_ port.PlayerSource  = (*Source)(nil)
)

func New(layout Layout, attach Attach) *Source {
	return &Source{layout: layout, attach: attach, now: time.Now}
}

// DuckStation is MR2 running under DuckStation. It reads memory only on Windows.
func DuckStation() *Source { return New(ps1, attachDuckStation) }

// MR2DX is the English release of Monster Rancher 1&2 DX from Steam, playing MR2. It reads memory only on Windows.
func MR2DX() *Source { return New(steam, attachMR2DX) }

// Current is the monster in the game's RAM now. When the game cannot be reached the error is apperr.ErrNotFound
// and says why, because to the viewer that is the same as the game showing no monster.
func (s *Source) Current(_ context.Context) (monster.Monster, error) {
	var m monster.Monster
	err := s.read(func(mem Memory) (err error) {
		m, err = decode(s.layout, mem)
		return err
	})
	return m, err
}

// Player is the player's side of the game in its RAM now, with the same errors as Current.
func (s *Source) Player(_ context.Context) (player.Player, error) {
	var p player.Player
	err := s.read(func(mem Memory) (err error) {
		p, err = decodePlayer(s.layout, mem)
		return err
	})
	return p, err
}

// isAttached is whether the game is being read now, which is how Auto tells a game that shows no monster from a
// game that is not there.
func (s *Source) isAttached() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mem != nil
}

// read runs fn on the game's RAM, attaching first when it must. When fn fails for any reason but that no
// monster is loaded, the game is let go of, so that the next call attaches again.
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
