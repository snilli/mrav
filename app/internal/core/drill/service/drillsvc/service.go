package drillsvc

import (
	"context"

	"app/internal/core/drill"
	"app/internal/core/drill/domain"
	"app/internal/port"
)

type Service struct {
	repo   drill.DrillRepository
	source port.MonsterSource
}

func New(repo drill.DrillRepository, source port.MonsterSource) drill.DrillService {
	return &Service{repo: repo, source: source}
}

func (s *Service) Current(ctx context.Context) (domain.Training, error) {
	m, err := s.source.Current(ctx)
	if err != nil {
		return domain.Training{}, err
	}
	cells, err := s.repo.ListGains(ctx)
	if err != nil {
		return domain.Training{}, err
	}
	return domain.Build(m, domain.NewGains(cells)), nil
}
