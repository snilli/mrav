package trainingsvc

import (
	"context"

	"app/internal/core/training"
	"app/internal/core/training/domain"
	"app/internal/port"
)

type Service struct {
	repo   training.TrainingRepository
	source port.MonsterSource
}

func New(repo training.TrainingRepository, source port.MonsterSource) training.TrainingService {
	return &Service{repo: repo, source: source}
}

func (s *Service) Current(ctx context.Context) (domain.Training, error) {
	m, err := s.source.Current(ctx)
	if err != nil {
		return domain.Training{}, err
	}
	gains, err := s.repo.ListGains(ctx)
	if err != nil {
		return domain.Training{}, err
	}
	return domain.Build(m, domain.NewGains(gains)), nil
}
