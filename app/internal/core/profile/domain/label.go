// Package domain is how the game's numbers read on the Monster, Life Index and Training pages: grades, labels,
// names, and the rules MR2AV applies to them. Pure Go.
package domain

import (
	"fmt"

	monster "app/internal/core/monster/domain"
)

// Level is how a percentage reads.
type Level int

const (
	LevelNone Level = iota
	LevelLow
	LevelMid
	LevelHigh
)

// NatureLabel is what an effective nature is called.
type NatureLabel string

const (
	Neutral NatureLabel = "Neutral"
	Bad     NatureLabel = "Bad"
	Worst   NatureLabel = "Worst"
	Good    NatureLabel = "Good"
	Best    NatureLabel = "Best"
)

var growthGrades = [...]string{"E", "D", "C", "B", "A", "S", "S2", "S3", "S4", "S5", "F"}

// GrowthGrade is the letter of a growth rate: 0 is E up to 4 for A, and S and F lie beyond.
func GrowthGrade(rate int) string {
	if rate < 0 || rate >= len(growthGrades) {
		return "---"
	}
	return growthGrades[rate]
}

// ArenaGrade is the letter of the arena speed: 0 is E up to 4 for A.
func ArenaGrade(speed int) string {
	if speed < 0 || speed > 4 {
		return "---"
	}
	return growthGrades[speed]
}

// NatureOf names an effective nature.
func NatureOf(effective int) NatureLabel {
	switch {
	case effective <= -60:
		return Worst
	case effective <= -20:
		return Bad
	case effective < 20:
		return Neutral
	case effective < 60:
		return Good
	default:
		return Best
	}
}

// FormLabel describes a form, which runs from -100 (skinny) to 100 (plump).
func FormLabel(form int) string {
	switch {
	case form >= -19 && form <= 19:
		return "Normal"
	case form >= 20 && form <= 59:
		return "Fat"
	case form >= 60 && form <= 100:
		return "Plump"
	case form >= -59 && form <= -20:
		return "Slim"
	case form >= -100 && form <= -60:
		return "Skinny"
	}
	return "---"
}

// MotivationOf says how eager a monster is for a drill: under 30 percent is low, 81 and over is high.
func MotivationOf(percent int) Level {
	switch {
	case percent > -1 && percent < 30:
		return LevelLow
	case percent < 81:
		return LevelMid
	case percent < 101:
		return LevelHigh
	}
	return LevelNone
}

var lifeStages = [...]string{
	"1 - Baby", "2 - Child", "3 - Adolescent", "4 - Adolescent 2", "5 - Prime",
	"6 - Sub-Prime", "7 - Elder", "8 - Elder 2", "9 - Old Age", "10 - Twilight",
}

// LifeStageText names a life stage, and says when the monster has no lifespan left, whatever its stage.
func LifeStageText(stage, lifespan int) string {
	switch {
	case lifespan <= 0:
		return "Dead :("
	case stage < 0 || stage >= len(lifeStages):
		return "---"
	}
	return lifeStages[stage]
}

// LifeTypeName names the way a monster ages.
func LifeTypeName(lifeType int) string {
	switch lifeType {
	case 0:
		return "Type 1: Normal"
	case 1:
		return "Type 2: Precocious"
	case 2:
		return "Type 3: Late Bloom"
	case 3:
		return "Type 4: Sustainable"
	case 4:
		return "Type 5: Prodigy [X]"
	}
	return "Unrecognised"
}

// PlayTypeName names the way a monster likes to play.
func PlayTypeName(play int) string {
	switch play {
	case 0:
		return "Mud Fight"
	case 1:
		return "Sumo Battle"
	case 2:
		return "Sparring"
	}
	return "---"
}

// GenusLabel is a genus's name. The codes 38 to 43 are breeds that are not in the game's list and show as
// "??? #n" as a sub breed; as a main breed they, and any other code without a name, show as "---".
func GenusLabel(code int, main bool, names map[monster.Genus]string) string {
	if name, ok := names[monster.Genus(code)]; ok {
		return name
	}
	if !main && code >= 38 && code <= 43 {
		return fmt.Sprintf("??? #%d", code-37)
	}
	return "---"
}
