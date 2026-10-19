package movesvc

import (
	"context"

	"app/internal/core/move/domain"
)

func (s *Service) ListGenera(ctx context.Context) ([]domain.Genus, error) {
	return s.repo.ListGenera(ctx)
}
