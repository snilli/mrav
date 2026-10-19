package movesvc

import (
	"context"

	monster "app/internal/core/monster/domain"
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
	genera, err := s.repo.ListGenera(ctx)
	if err != nil {
		return domain.Grid{}, err
	}
	names := make(map[monster.Genus]string, len(genera))
	for _, g := range genera {
		names[g.Code] = g.Name
	}
	return domain.NewGridNamed(genus, moves, m, domain.Names{Genus: func(g monster.Genus) string { return names[g] }}), nil
}
