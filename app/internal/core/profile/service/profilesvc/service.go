package profilesvc

import (
	"context"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile"
	"app/internal/core/profile/domain"
	"app/internal/port"
)

type Service struct {
	repo     profile.ProfileRepository
	monsters port.MonsterSource
	players  port.PlayerSource
}

func New(repo profile.ProfileRepository, monsters port.MonsterSource, players port.PlayerSource) profile.ProfileService {
	return &Service{repo: repo, monsters: monsters, players: players}
}

func (s *Service) Current(ctx context.Context) (domain.Profile, error) {
	m, err := s.monsters.Current(ctx)
	if err != nil {
		return domain.Profile{}, err
	}
	p, err := s.players.Player(ctx)
	if err != nil {
		return domain.Profile{}, err
	}
	genera, err := s.repo.ListGenusNames(ctx)
	if err != nil {
		return domain.Profile{}, err
	}
	rules, err := s.repo.ListBreedRules(ctx, int(m.Genus))
	if err != nil {
		return domain.Profile{}, err
	}
	cocoons, err := s.repo.ListCocoonRules(ctx)
	if err != nil {
		return domain.Profile{}, err
	}
	desires, err := s.repo.ListDesireNames(ctx)
	if err != nil {
		return domain.Profile{}, err
	}
	// The worth of a search of each place of the maps is the database's to work out, for how this monster searches.
	maps, err := s.repo.ListExpeditionMaps(ctx, domain.Searcher{
		Intelligence: m.Stats[monster.Int], Skill: m.Stats[monster.Skill], GoodNature: m.Nature > 0,
	})
	if err != nil {
		return domain.Profile{}, err
	}

	drug := "---"
	if domain.IsDrug(m.Drug) {
		drug = domain.DesireName(desires, m.Drug)
	}
	var breeds []domain.BreedUnlock
	if p.Ranch.Known {
		breeds = domain.BreedUnlocks(genera, p.Ranch.Unlocked)
	}
	return domain.Profile{
		Breeds:    breeds,
		Maps:      maps,
		Monster:   m,
		Player:    p,
		Breed:     domain.ResolveBreed(rules, cocoons, m),
		MainGenus: domain.GenusLabel(int(m.Genus), true, genera),
		SubGenus:  domain.GenusLabel(int(m.SubGenus), false, genera),
		Like:      domain.DesireName(desires, m.ItemLike),
		Dislike:   domain.DesireName(desires, m.ItemDislike),
		Drug:      drug,
	}, nil
}
