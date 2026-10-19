package movesvc

import (
	"context"

	"app/internal/core/move/domain"
)

func (s *Service) CurrentGrid(ctx context.Context) (domain.Grid, error) {
	m, err := s.source.Current(ctx)
	if err != nil {
		return domain.Grid{}, err
	}
	genus, err := s.repo.GetGenus(ctx, m.Genus)
	if err != nil {
		return domain.Grid{}, err
	}
	moves, err := s.repo.ListMoves(ctx, m.Genus)
	if err != nil {
		return domain.Grid{}, err
	}
	return domain.NewGrid(genus, moves, m), nil
}
