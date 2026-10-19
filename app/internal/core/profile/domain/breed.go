package domain

import (
	"slices"

	monster "app/internal/core/monster/domain"
)

const (
	// NoMonster is the name of a main breed the rules know nothing of.
	NoMonster = "No Monster"
	// UndefinedMonster is the name of a sub breed that the rules of its main breed know nothing of.
	UndefinedMonster = "Undefined Monster"

	protomessiah = "Protomessiah" // a Henger and Gali that has been revived
	wormSub      = 36
)

// realBeaclon is the growth rates of a Beaclon that was born one, which a cocooned Worm's are not.
var realBeaclon = [monster.StatCount]int{3, 3, 0, 1, 1, 3}

// BreedRule names a breed, a main and a sub genus together, from the monster's growth rates. A breed has one
// rule, or several when it comes in variants, which are tried in order.
type BreedRule struct {
	Main, Sub int
	Name      string // when the rule matches
	Otherwise string // the name when no rule of the breed matches, usually Name with "(N/S)"; empty means Name
	Guts      *int   // the breed's guts rate, which the guts box checks, and which a variant can match on
	MatchGuts bool   // the rule matches only when the monster has that guts rate
	Growth    *[monster.StatCount]int
}

func (r BreedRule) matches(m monster.Monster) bool {
	if r.Growth != nil && *r.Growth != m.Growth {
		return false
	}
	return !r.MatchGuts || r.Guts != nil && *r.Guts == m.GutsRate
}

// CocoonRule names a cocooned Worm from its guts rate and growth rates.
type CocoonRule struct {
	Guts   int
	Growth [monster.StatCount]int
	Name   string
}

// Breed is a monster's breed name, and whether its guts rate is the one that breed has.
type Breed struct {
	Name   string
	GutsOK bool
}

// ResolveBreed names the breed of a monster. The rules are those of its main breed. The order is MR2AV's: a
// revived Henger and Gali first, then a cocoon, then the rules of the breed.
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

	// The guts rate is right when it is one the breed has; a breed that has none to check is right from 6 up,
	// because no monster has less.
	gutsOK := m.GutsRate >= 6
	if slices.ContainsFunc(forSub, func(r BreedRule) bool { return r.Guts != nil }) {
		gutsOK = slices.ContainsFunc(forSub, func(r BreedRule) bool { return r.Guts != nil && *r.Guts == m.GutsRate })
	}

	for _, r := range forSub {
		if r.matches(m) {
			return Breed{Name: r.Name, GutsOK: gutsOK}
		}
	}
	last := forSub[len(forSub)-1]
	if last.Otherwise == "" {
		return Breed{Name: last.Name, GutsOK: gutsOK}
	}
	return Breed{Name: last.Otherwise, GutsOK: gutsOK}
}

// PossibleCocoon says whether a monster might be a cocooned Worm: a Beaclon of Beaclon whose growth rates are
// not a real one's, or a Pixie, Hare, Gali, Suezo, Jell, Monol or Naga with Worm for a sub breed.
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

// ShowsCocoon says whether the cocoon facts matter: for a Worm of rank C or lower, which can still cocoon.
func ShowsCocoon(m monster.Monster) bool { return m.Genus == 36 && m.Rank < 3 }

// CocoonReady says whether a Worm will cocoon: four years old, 80 loyalty, no fatigue and under 30 stress.
func CocoonReady(m monster.Monster) bool {
	return m.Age >= 192 && Loyalty(m.Spoil, m.Fear) >= 80 && m.Fatigue == 0 && m.Stress < 30
}
