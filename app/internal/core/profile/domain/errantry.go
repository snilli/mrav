package domain

import (
	"slices"

	monster "app/internal/core/monster/domain"
)

// ErrantryWeeks is how many weeks an errantry lasts, each with a chance of its own to succeed.
const ErrantryWeeks = 4

// errantryFatiguePerWeek is how much fatigue an errantry adds each week the monster is away.
const errantryFatiguePerWeek = 18

// errantryThresholds are what a week asks of the monster's strength: the later the week, the more.
var errantryThresholds = [ErrantryWeeks]int{5, 15, 25, 35}

// ErrantryLocale is a place that a monster can be sent to, with the stat that the trip trains.
type ErrantryLocale struct {
	Name string
	Stat monster.Stat
}

// errantryLocales are the places of an errantry, in the order of the game.
var errantryLocales = [...]ErrantryLocale{
	{"Papas Mountain", monster.Speed},
	{"Mandy Desert", monster.Pow},
	{"Parepare Jungle", monster.Int},
	{"Torble Sea", monster.Skill},
	{"Kawrea Volcano", monster.Defense},
}

// ErrantryGain is what an errantry gives: the stat of the place, from Min to Max, and up to Life of Life.
type ErrantryGain struct{ Min, Max, Life int }

// errantryGains are the gains by life stage, from the first, as the page gives them: the stat from its least to its most for
// each stat gain from 1 to 5, and then the Life that goes with it, which the game keeps under the gain of the stat.
var errantryGains = [10]struct {
	stat [5][2]int
	life int
}{
	{[5][2]int{{2, 4}, {2, 5}, {2, 5}, {2, 6}, {3, 7}}, 3},
	{[5][2]int{{2, 5}, {2, 6}, {3, 7}, {4, 8}, {5, 9}}, 5},
	{[5][2]int{{3, 7}, {4, 8}, {5, 9}, {6, 10}, {8, 12}}, 7},
	{[5][2]int{{3, 7}, {4, 8}, {5, 9}, {6, 10}, {8, 12}}, 7},
	{[5][2]int{{5, 9}, {7, 11}, {8, 12}, {10, 14}, {13, 15}}, 10},
	{[5][2]int{{4, 8}, {6, 10}, {7, 11}, {9, 13}, {11, 15}}, 9},
	{[5][2]int{{3, 7}, {5, 9}, {6, 10}, {7, 11}, {9, 13}}, 8},
	{[5][2]int{{3, 7}, {4, 8}, {5, 9}, {6, 10}, {8, 12}}, 7},
	{[5][2]int{{2, 5}, {2, 6}, {3, 7}, {4, 8}, {5, 9}}, 5},
	{[5][2]int{{2, 4}, {2, 5}, {2, 5}, {2, 6}, {3, 7}}, 3},
}

// ErrantryGainOf is what an errantry gives at a life stage, from 1 to 10, to a stat whose gain is from 1 to 5.
func ErrantryGainOf(stage, level int) ErrantryGain {
	row := errantryGains[min(max(stage, 1), 10)-1]
	stat := row.stat[min(max(level, 1), 5)-1]
	return ErrantryGain{Min: stat[0], Max: stat[1], Life: row.life}
}

// ErrantryLocales are the places of an errantry, in the order of the game, as a copy that is safe to change.
func ErrantryLocales() []ErrantryLocale { return slices.Clone(errantryLocales[:]) }

// ErrantryChance is the chance, in percent, that a week of an errantry succeeds: 29 rolls, of which as many succeed as the
// monster's strength is above what the week asks. The strength is twice the third of twice the stat gain of the place and
// the life gain, 4 for each life stage, the fear over 5, less the fatigue over 5, which is 18 more for each week gone. The
// stage runs from 1 to 10, the week from 1 to 4, and the gains from 1 (the growth rate E) to 5 (A). It is the calculator of
// legendcup.com, which has it from SmilingFaces96.
func ErrantryChance(stage, week, statGain, lifeGain, fatigue, fear int) float64 {
	fatigue += errantryFatiguePerWeek * (week - 1)
	strength := 2*((2*statGain+lifeGain)/3) + 4*stage + fear/5 - fatigue*2/10
	chance := float64(28-errantryThresholds[week-1]+strength) / 29 * 100
	return min(max(chance, 0), 100)
}

// ErrantryOutlook is how a monster stands at one place: the gain of its stat there, what the trip would give it, and the
// chance of each week.
type ErrantryOutlook struct {
	Locale  ErrantryLocale
	Level   int // 1 to 5, the growth rate of the place's stat from E to A
	Gains   ErrantryGain
	Chances [ErrantryWeeks]float64
}

// PlanErrantry works out, for each place, the chance of each week for the monster as it is now, as if it were sent at
// once. What the monster has is kept to what the calculator takes.
func PlanErrantry(m monster.Monster) []ErrantryOutlook {
	gain := func(stat monster.Stat) int { return min(max(m.Growth[stat]+1, 1), 5) }
	stage := min(max(m.LifeStage+1, 1), 10)
	fatigue, fear := min(max(m.Fatigue, 0), 100), min(max(m.Fear, 0), 100)

	outlook := make([]ErrantryOutlook, len(errantryLocales))
	for i, l := range errantryLocales {
		o := ErrantryOutlook{Locale: l, Level: gain(l.Stat)}
		o.Gains = ErrantryGainOf(stage, o.Level)
		for week := 1; week <= ErrantryWeeks; week++ {
			o.Chances[week-1] = ErrantryChance(stage, week, o.Level, gain(monster.Life), fatigue, fear)
		}
		outlook[i] = o
	}
	return outlook
}
