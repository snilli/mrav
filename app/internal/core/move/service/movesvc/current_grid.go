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

// keepToErrantry marks ready only the moves that a place of the errantry would teach now, with a chance the Errantry page
// shows, so that the Moves page counts what the Errantry page lists: a move whose rule is met is not ready while one before
// it at its place takes all the chance, and one that the place may teach by luck before its rule is met is. A move held back
// so is given one more requirement, which names the move that its place teaches first.
func keepToErrantry(g *domain.Grid, m monster.Monster, moves []domain.Move) {
	taught := map[int]bool{}  // by memory slot
	first := map[int]string{} // by memory slot: what a place teaches before a move that it holds back
	locales := profile.ErrantryLocales()
	for i, place := range profile.LearnChances(m, moves) {
		for j, l := range place {
			slot := l.Move.MemorySlot
			taught[slot] = taught[slot] || l.Shown()
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
			cell.CanLearn = taught[slot]
			if !taught[slot] && first[slot] != "" {
				cell.Requirements = append(cell.Requirements, domain.Requirement{Text: first[slot]})
			}
		}
	}
}
