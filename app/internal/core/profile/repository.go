package profile

import (
	"context"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
)

type ProfileRepository interface {
	ListGenusNames(ctx context.Context) (map[monster.Genus]string, error)
	ListBreedRules(ctx context.Context, main int) ([]domain.BreedRule, error)
	ListCocoonRules(ctx context.Context) ([]domain.CocoonRule, error)
	ListDesireNames(ctx context.Context) (map[int]string, error)
	ListExpeditionMaps(ctx context.Context) ([]domain.ExpeditionMap, error)
}
