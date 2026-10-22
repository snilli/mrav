package domain

import (
	"slices"

	monster "app/internal/core/monster/domain"
)

const ErrantryWeeks = 4

const errantryFatiguePerWeek = 18

var errantryThresholds = [ErrantryWeeks]int{5, 15, 25, 35}

type ErrantryLocale struct {
	Name string
	Stat monster.Stat
}

var errantryLocales = [...]ErrantryLocale{
	{"Papas Mountain", monster.Speed},
	{"Mandy Desert", monster.Pow},
	{"Parepare Jungle", monster.Int},
	{"Torble Sea", monster.Skill},
	{"Kawrea Volcano", monster.Defense},
}

type ErrantryGain struct{ Min, Max, Life int }

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

func ErrantryGainOf(stage, level int) ErrantryGain {
	row := errantryGains[min(max(stage, 1), 10)-1]
	stat := row.stat[min(max(level, 1), 5)-1]
	return ErrantryGain{Min: stat[0], Max: stat[1], Life: row.life}
}

func ErrantryLocales() []ErrantryLocale { return slices.Clone(errantryLocales[:]) }

func ErrantryChance(stage, week, statGain, lifeGain, fatigue, fear int) float64 {
	fatigue += errantryFatiguePerWeek * (week - 1)
	strength := 2*((2*statGain+lifeGain)/3) + 4*stage + fear/5 - fatigue*2/10
	chance := float64(28-errantryThresholds[week-1]+strength) / 29 * 100
	return min(max(chance, 0), 100)
}

type ErrantryOutlook struct {
	Locale  ErrantryLocale
	Level   int
	Gains   ErrantryGain
	Chances [ErrantryWeeks]float64
}

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
