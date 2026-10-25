package domain

import (
	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
)

type Profile struct {
	Monster   monster.Monster
	Player    player.Player
	Breed     Breed
	MainGenus string
	SubGenus  string
	Like      string
	Dislike   string
	Drug      string
	Breeds    []BreedUnlock
	Maps      []ExpeditionMap
	Calendar  []CalendarEntry
}

func DesireName(names map[int]string, id int) string {
	if name, ok := names[id]; ok {
		return name
	}
	return "---"
}
