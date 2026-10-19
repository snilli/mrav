package gamememory

import (
	"context"
	"errors"
	"strings"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/port"
)

// Auto reads whichever of several games is running, so that one build suits a player of any of them. It asks the
// games in order and takes the first it is attached to; a game that is attached but shows no monster still wins,
// because that is the game being played.
type Auto struct{ games []*Source }

var (
	_ port.MonsterSource = (*Auto)(nil)
	_ port.PlayerSource  = (*Auto)(nil)
)

func NewAuto(games ...*Source) *Auto { return &Auto{games: games} }

// AutoDetect is whichever of DuckStation and MR2DX is running. It reads memory only on Windows.
func AutoDetect() *Auto { return NewAuto(DuckStation(), MR2DX()) }

func (a *Auto) Current(ctx context.Context) (monster.Monster, error) {
	return firstGame(a.games, func(s *Source) (monster.Monster, error) { return s.Current(ctx) })
}

func (a *Auto) Player(ctx context.Context) (player.Player, error) {
	return firstGame(a.games, func(s *Source) (player.Player, error) { return s.Player(ctx) })
}

// firstGame reads from the first game that can be reached. When none can, the error is apperr.ErrNotFound and
// gives the reason of each.
func firstGame[T any](games []*Source, read func(*Source) (T, error)) (T, error) {
	var reasons []string
	for _, g := range games {
		v, err := read(g)
		if err == nil || g.isAttached() {
			return v, err
		}
		reasons = append(reasons, reasonOf(err))
	}
	var none T
	return none, apperr.NotFound("game", strings.Join(reasons, "; "))
}

// reasonOf is what a not-found error says, or what any other error says.
func reasonOf(err error) string {
	if e, ok := errors.AsType[*apperr.Error](err); ok {
		return e.Msg
	}
	return err.Error()
}
