package profilesvc

import (
	"cmp"
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
	version := cmp.Or(m.Version, monster.MR2DX)
	rules, err := s.repo.ListBreedRules(ctx, int(m.Genus), version)
	if err != nil {
		return domain.Profile{}, err
	}
	cocoons, err := s.repo.ListCocoonRules(ctx, version)
	if err != nil {
		return domain.Profile{}, err
	}
	desires, err := s.repo.ListDesireNames(ctx)
	if err != nil {
		return domain.Profile{}, err
	}
	shared, err := s.repo.ListExpeditionMaps(ctx)
	if err != nil {
		return domain.Profile{}, err
	}
	by := domain.Searcher{Intelligence: m.Stats[monster.Int], Skill: m.Stats[monster.Skill], GoodNature: m.Nature > 0}
	var maps []domain.ExpeditionMap
	for _, sm := range shared {
		maps = append(maps, sm.For(by))
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
