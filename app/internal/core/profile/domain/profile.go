package domain

import (
	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
)

// Profile is what the Monster page shows: the monster and the player as the game has them, and the names that
// come from the game's data rather than from its memory.
type Profile struct {
	Monster   monster.Monster
	Player    player.Player
	Breed     Breed
	MainGenus string
	SubGenus  string
	Like      string        // the item the monster wants
	Dislike   string        // the item the monster does not want
	Drug      string        // the item that is a drug the monster is on; "---" when it is on none
	Breeds    []BreedUnlock // the main breeds and whether each is unlocked; none when the game does not say
}

// DesireName is what an item id is called when a monster wants it, dislikes it or is on it, or "---" for an id
// the list does not have.
func DesireName(names map[int]string, id int) string {
	if name, ok := names[id]; ok {
		return name
	}
	return "---"
}
