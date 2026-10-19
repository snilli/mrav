package domain

import (
	"fmt"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
)

// BreedUnlock is a main breed and whether the game has unlocked it for the breeder to raise.
type BreedUnlock struct {
	Name     string
	Unlocked bool
}

// BreedUnlocks lists the main breeds in the game's order, each with whether it is unlocked. A breed that the names
// lack is called by its code.
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
