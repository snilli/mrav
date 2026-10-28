package plannersvc

import (
	"context"

	"app/internal/core/planner"
	"app/internal/core/planner/domain"
	"app/internal/port"
)

type Service struct {
	repo     planner.PlannerRepository
	monsters port.MonsterSource
	players  port.PlayerSource
}

func New(repo planner.PlannerRepository, monsters port.MonsterSource, players port.PlayerSource) planner.PlannerService {
	return &Service{repo: repo, monsters: monsters, players: players}
}

func (s *Service) Current(ctx context.Context) (domain.Current, error) {
	m, err := s.monsters.Current(ctx)
	if err != nil {
		return domain.Current{}, err
	}
	p, err := s.players.Player(ctx)
	if err != nil {
		return domain.Current{}, err
	}
	reactions, err := s.repo.ListFoodReactions(ctx)
	if err != nil {
		return domain.Current{}, err
	}
	effects, err := s.repo.ListEffects(ctx)
	if err != nil {
		return domain.Current{}, err
	}
	return domain.Current{Settings: domain.SettingsOf(m, p.Items), Reactions: domain.NewReactions(reactions), Effects: domain.NewEffects(effects)}, nil
}
