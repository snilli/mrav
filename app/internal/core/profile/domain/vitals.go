package domain

import (
	"fmt"
	"strconv"
)

type LifeIndex struct {
	Value     int
	WeeksLost int
}

func LifeIndexOf(fatigue, stress int) LifeIndex {
	value := fatigue + stress*2
	weeks := 0
	for _, above := range []int{69, 104, 139, 174, 209, 244, 279} {
		if value > above {
			weeks++
		}
	}
	return LifeIndex{Value: value, WeeksLost: weeks}
}

func clampStat(v float32) float32 { return max(min(v, 999), 1) }

func EffectiveDefense(def, form int) int {
	return int(clampStat(float32(def) * (1 + float32(form)/400)))
}

func EffectiveSpeed(spd, form int) int {
	return int(clampStat(float32(spd) * (1 - float32(form)/400)))
}

func FatigueWord(fatigue int) string {
	switch {
	case fatigue >= 100:
		return "pass out"
	case fatigue >= 80:
		return "very tired"
	case fatigue >= 60:
		return "pretty tired"
	case fatigue >= 40:
		return "seems tired"
	case fatigue >= 20:
		return "seems well"
	case fatigue >= 1:
		return "well"
	}
	return "very well"
}

func Loyalty(spoil, fear int) int { return spoil/2 + fear/2 }

func AgeText(weeks int) string {
	if weeks == -1 {
		return "---"
	}
	return fmt.Sprintf("%dy,%dm,%dw", weeks/48, weeks%48/4, weeks%4)
}

func GutsText(guts int, perSecond bool) string {
	if guts >= 50 {
		if perSecond {
			return "30.00"
		}
		return fmt.Sprintf("%d (1)", guts)
	}
	if !perSecond {
		return strconv.Itoa(guts)
	}
	var gutsPerSecond float32
	if guts > 0 {
		gutsPerSecond = 30 / float32(guts)
	}
	return fmt.Sprintf("%.2f", gutsPerSecond)
}
