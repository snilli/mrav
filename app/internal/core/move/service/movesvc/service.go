package movesvc

import (
	"app/internal/core/move"
	"app/internal/port"
)

type Service struct {
	repo   move.MoveRepository
	source port.MonsterSource
}

func New(repo move.MoveRepository, source port.MonsterSource) move.MoveService {
	return &Service{repo: repo, source: source}
}
