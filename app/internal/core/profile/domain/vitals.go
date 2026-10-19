package domain

import (
	"fmt"
	"strconv"
)

// LifeIndex is what fatigue and stress add up to, and how many weeks of life that takes from the monster.
type LifeIndex struct {
	Value     int
	WeeksLost int
}

// LifeIndexOf adds the fatigue and twice the stress. Every 35 above 69 costs a week, up to 7.
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

// The effective stats are single precision, as in MR2AV, because a whole number that is a hair under is cut down.
func clampStat(v float32) float32 { return max(min(v, 999), 1) }

// EffectiveDefense is the defense once the form has changed it: a fat monster defends better. It is a whole
// number from 1 to 999.
func EffectiveDefense(def, form int) int {
	return int(clampStat(float32(def) * (1 + float32(form)/400)))
}

// EffectiveSpeed is the speed once the form has changed it: a fat monster is slower.
func EffectiveSpeed(spd, form int) int {
	return int(clampStat(float32(spd) * (1 - float32(form)/400)))
}

// Loyalty is half of the spoil and half of the fear, each cut to a whole number, added.
func Loyalty(spoil, fear int) int { return spoil/2 + fear/2 }

// AgeText says an age in weeks as years, months and weeks, of a year of 48 weeks and a month of 4.
func AgeText(weeks int) string {
	if weeks == -1 {
		return "---"
	}
	return fmt.Sprintf("%dy,%dm,%dw", weeks/48, weeks%48/4, weeks%4)
}

// GutsText shows the guts rate as the game keeps it, or as guts a second. A rate of 50 or more is one guts in
// 30 seconds, which the game shows as "(1)", and 30 guts a second.
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
