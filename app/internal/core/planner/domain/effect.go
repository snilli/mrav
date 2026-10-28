package domain

import (
	"fmt"
	"strings"

	monster "app/internal/core/monster/domain"
)

var (
	itemKeys = [itemCount]string{
		"none", "mint_leaf", "nuts_oil", "magic_banana_plus", "magic_banana_zero", "magic_banana_minus", "irritater", "mango", "paradoxine", "troron",
		"candy", "smoked_snake", "apple_cake", "powder", "sweet_jelly", "sour_jelly", "star_prune", "gold_peach", "silver_peach", "griever",
	}
	actionKeys = map[Action]string{
		Rest: "rest", LightDrill: "light_drill", HardDrill: "hard_drill", Errantry: "errantry",
		BattleWin: "battle_win", BattleOK: "battle_ok", BattleLose: "battle_lose", Expedition: "expedition",
	}
)

type Effect struct {
	Subject      string
	Version      monster.Version
	Genus        *int
	LifeStage    *int
	StressFactor *float64
	Fatigue      int
	Stress       int
	FormDivisor  *int
	Lost         int
	Form         int
}

func (f Effect) apply(ftg, str, form int) (int, int) {
	if f.StressFactor != nil {
		str = ceil(float64(str) * *f.StressFactor)
	}
	ftg, str = ftg+f.Fatigue, str+f.Stress
	if f.FormDivisor != nil {
		ftg += form / *f.FormDivisor
	}
	return ftg, str
}

type effectKey struct {
	subject string
	version monster.Version
}

type Effects struct{ bySubject map[effectKey][]Effect }

func NewEffects(list []Effect) Effects {
	e := Effects{bySubject: make(map[effectKey][]Effect, len(list))}
	for _, f := range list {
		k := effectKey{f.Subject, f.Version}
		e.bySubject[k] = append(e.bySubject[k], f)
	}
	return e
}

func (e Effects) of(subject string, s Settings) Effect {
	var anyone Effect
	for _, f := range e.bySubject[effectKey{subject, s.Version}] {
		switch {
		case f.Genus != nil:
			if *f.Genus == s.Main || *f.Genus == s.Sub {
				return f
			}
		case f.LifeStage != nil:
			if *f.LifeStage == s.LifeStage {
				return f
			}
		default:
			anyone = f
		}
	}
	return anyone
}

func (e Effects) Text(item Item, s Settings) string {
	f := e.of(itemSubject(item), s)
	var parts []string
	if f.Fatigue != 0 {
		parts = append(parts, fmt.Sprintf("Fatigue %+d", f.Fatigue))
	}
	if f.StressFactor != nil {
		parts = append(parts, fmt.Sprintf("Stress x%g", *f.StressFactor))
	}
	if f.Stress != 0 {
		parts = append(parts, fmt.Sprintf("Stress %+d", f.Stress))
	}
	if f.Lost != 0 {
		unit := "weeks"
		if f.Lost == 1 || f.Lost == -1 {
			unit = "week"
		}
		parts = append(parts, fmt.Sprintf("Life %+d %s", f.Lost, unit))
	}
	if f.Form != 0 {
		parts = append(parts, fmt.Sprintf("Form %+d", f.Form))
	}
	if len(parts) == 0 {
		return "no change"
	}
	return strings.Join(parts, ", ")
}

func itemSubject(i Item) string { return "item:" + itemKeys[i] }

func foodSubject(f Food) string { return "food:" + f.Key() }

func actionSubject(a Action) string { return "action:" + actionKeys[a] }

func KnownSubject(subject string) bool {
	for _, i := range Items()[1:] {
		if itemSubject(i) == subject {
			return true
		}
	}
	for _, f := range Foods() {
		if foodSubject(f) == subject {
			return true
		}
	}
	for a := range actionKeys {
		if actionSubject(a) == subject {
			return true
		}
	}
	return false
}

func clampForm(v int) int { return max(-100, min(100, v)) }
