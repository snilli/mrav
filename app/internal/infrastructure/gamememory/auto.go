package gamememory

import (
	"context"
	"errors"
	"runtime"
	"strings"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/port"
)

type Auto struct{ games []*Source }

var (
	_ port.MonsterSource = (*Auto)(nil)
	_ port.PlayerSource  = (*Auto)(nil)
)

func NewAuto(games ...*Source) *Auto { return &Auto{games: games} }

func AutoDetect() *Auto {
	if runtime.GOOS == "darwin" {
		return NewAuto(DuckStation())
	}
	return NewAuto(DuckStation(), MR2DX())
}

func (a *Auto) Current(ctx context.Context) (monster.Monster, error) {
	return firstGame(a.games, func(s *Source) (monster.Monster, error) { return s.Current(ctx) })
}

func (a *Auto) Player(ctx context.Context) (player.Player, error) {
	return firstGame(a.games, func(s *Source) (player.Player, error) { return s.Player(ctx) })
}

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

func reasonOf(err error) string {
	if e, ok := errors.AsType[*apperr.Error](err); ok {
		return e.Msg
	}
	return err.Error()
}
