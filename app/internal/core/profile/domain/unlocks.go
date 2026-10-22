package domain

import (
	"fmt"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
)

type BreedUnlock struct {
	Name     string
	Unlocked bool
}

func BreedUnlocks(names map[monster.Genus]string, unlocked [player.BreedCount]bool) []BreedUnlock {
	list := make([]BreedUnlock, player.BreedCount)
	for code, on := range unlocked {
		name, ok := names[monster.Genus(code)]
		if !ok {
			name = fmt.Sprintf("#%d", code)
		}
		list[code] = BreedUnlock{Name: name, Unlocked: on}
	}
	return list
}
