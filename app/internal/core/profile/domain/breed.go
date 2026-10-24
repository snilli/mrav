package domain

import (
	"slices"

	monster "app/internal/core/monster/domain"
)

const (
	NoMonster        = "No Monster"
	UndefinedMonster = "Undefined Monster"

	protomessiah = "Protomessiah"
	notStandard  = " (N/S)"
	wormSub      = 36
)

var realBeaclon = [monster.StatCount]int{3, 3, 0, 1, 1, 3}

type BreedRule struct {
	Main, Sub int
	Name      string
	Guts      *int
	MatchGuts bool
	Growth    *[monster.StatCount]int
}

func (r BreedRule) matches(m monster.Monster) bool {
	if r.Growth != nil && *r.Growth != m.Growth {
		return false
	}
	return !r.MatchGuts || r.Guts != nil && *r.Guts == m.GutsRate
}

type CocoonRule struct {
	Guts   int
	Growth [monster.StatCount]int
	Name   string
}

type Breed struct {
	Name   string
	GutsOK bool
}

func ResolveBreed(rules []BreedRule, cocoons []CocoonRule, m monster.Monster) Breed {
	main, sub := int(m.Genus), int(m.SubGenus)
	if main == 5 && sub == 15 {
		return Breed{Name: protomessiah, GutsOK: true}
	}
	if PossibleCocoon(main, sub, m.Growth) {
		for _, c := range cocoons {
			if c.Guts == m.GutsRate && c.Growth == m.Growth {
				return Breed{Name: c.Name, GutsOK: true}
			}
		}
	}
	return nameFromRules(rules, m)
}

func nameFromRules(rules []BreedRule, m monster.Monster) Breed {
	if len(rules) == 0 {
		return Breed{Name: NoMonster, GutsOK: true}
	}
	var forSub []BreedRule
	for _, r := range rules {
		if r.Main == int(m.Genus) && r.Sub == int(m.SubGenus) {
			forSub = append(forSub, r)
		}
	}
	if len(forSub) == 0 {
		return Breed{Name: UndefinedMonster, GutsOK: true}
	}

	gutsOK := m.GutsRate >= 6
	if slices.ContainsFunc(forSub, func(r BreedRule) bool { return r.Guts != nil }) {
		gutsOK = slices.ContainsFunc(forSub, func(r BreedRule) bool { return r.Guts != nil && *r.Guts == m.GutsRate })
	}

	for _, r := range forSub {
		if r.matches(m) {
			return Breed{Name: r.Name, GutsOK: gutsOK}
		}
	}
	return Breed{Name: forSub[len(forSub)-1].Name + notStandard, GutsOK: gutsOK}
}

func PossibleCocoon(main, sub int, growth [monster.StatCount]int) bool {
	if main == 4 && sub == 4 {
		return growth != realBeaclon
	}
	if sub != wormSub {
		return false
	}
	switch main {
	case 0, 13, 15, 23, 28, 34, 37:
		return true
	}
	return false
}

func ShowsCocoon(m monster.Monster) bool { return m.Genus == 36 && m.Rank < 3 }

func CocoonReady(m monster.Monster) bool {
	return m.Age >= 192 && Loyalty(m.Spoil, m.Fear) >= 80 && m.Fatigue == 0 && m.Stress < 30
}
