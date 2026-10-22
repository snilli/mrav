package movesvc

import (
	"context"
	"fmt"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
	profile "app/internal/core/profile/domain"
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
	grid := domain.NewGridNamed(genus, moves, m, domain.Names{Genus: func(g monster.Genus) string { return names[g] }})
	keepToErrantry(&grid, m, moves)
	return grid, nil
}

func keepToErrantry(g *domain.Grid, m monster.Monster, moves []domain.Move) {
	type teach struct {
		learn profile.Learn
		place string
	}
	best := map[int]teach{}
	first := map[int]string{}
	locales := profile.ErrantryLocales()
	for i, place := range profile.LearnChances(m, moves) {
		for j, l := range place {
			slot := l.Move.MemorySlot
			if l.Shown() && l.Chance > best[slot].learn.Chance {
				best[slot] = teach{l, locales[i].Name}
			}
			if !l.Shown() && j > 0 && first[slot] == "" {
				first[slot] = fmt.Sprintf("%s teaches %s first", locales[i].Name, place[0].Move.Name)
			}
		}
	}
	for c := range g.Cells {
		for r := range g.Cells[c] {
			cell := &g.Cells[c][r]
			if cell.Move == nil {
				continue
			}
			slot := cell.Move.MemorySlot
			t, taught := best[slot]
			cell.CanLearn, cell.Chance, cell.Place = taught && t.learn.Sure(), t.learn.Chance, t.place
			if !taught && first[slot] != "" {
				cell.Requirements = append(cell.Requirements, domain.Requirement{Text: first[slot]})
			}
		}
	}
}
