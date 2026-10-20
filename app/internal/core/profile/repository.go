// Package profile is the Monster page's use case: the monster the game shows, named and described from the game's data.
package profile

import (
	"context"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
)

// ProfileRepository is the read port of the data the Monster page names things from.
type ProfileRepository interface {
	// ListGenusNames returns the name of every genus, by its code.
	ListGenusNames(ctx context.Context) (map[monster.Genus]string, error)
	// ListBreedRules returns the rules that name the breeds of a main genus, ordered by sub genus and then by
	// their order within a breed; a main genus with none gives none.
	ListBreedRules(ctx context.Context, main int) ([]domain.BreedRule, error)
	// ListCocoonRules returns the rules that name a cocooned Worm.
	ListCocoonRules(ctx context.Context) ([]domain.CocoonRule, error)
	// ListDesireNames returns what each item id is called when a monster wants it, dislikes it or is on it.
	ListDesireNames(ctx context.Context) (map[int]string, error)
	// ListExpeditionMaps returns the maps of the trips, in the order they are kept, each with its places by letter;
	// what a search of a place is worth, in gold, is worked out by the database for a monster that searches as the
	// searcher does.
	ListExpeditionMaps(ctx context.Context, by domain.Searcher) ([]domain.ExpeditionMap, error)
}
