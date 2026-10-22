package domain

import (
	"cmp"

	monster "app/internal/core/monster/domain"
)

const (
	MoveCost     = 1
	ObstacleCost = 5
	SearchCost   = 10
	FailBelow    = -9
	FailWeeks    = 12
)

type Expedition struct {
	Name        string
	Host        string
	Month, Week int
	Life, Fame  int
	Rank        int
	Origin      Point
	RockPow     int
}

func (e Expedition) BreakChance(pow int) float64 {
	return min(max(0.5+float64(pow-cmp.Or(e.RockPow, 350))/300, 0.05), 1)
}

var Expeditions = []Expedition{
	{Name: "Torles Mountains", Host: "Rovest", Month: 2, Week: 1, Life: 210, Fame: 50, Rank: 2},
	{Name: "Kawrea Volcano", Host: "Dr. Talico", Month: 6, Week: 1, Life: 280, Fame: 50, Rank: 3},
	{Name: "Parepare Jungle", Host: "Kavaro", Month: 10, Week: 1, Life: 140, Fame: 40, Rank: 2},
}

var Phoenix = Expedition{
	Name: "Phoenix Volcano Adventure", Month: 6, Week: 1, Life: 140, Fame: 40, Rank: 1,
	Origin: Point{X: 8, Y: 14}, RockPow: 100,
}

type ConditionKind int

const (
	CondLife ConditionKind = iota
	CondFame
	CondGrade
)

type Condition struct {
	Kind       ConditionKind
	Need, Have int
}

func (c Condition) Met() bool { return c.Have >= c.Need }

func (e Expedition) Fit(m monster.Monster) bool {
	for _, c := range e.Conditions(m) {
		if !c.Met() {
			return false
		}
	}
	return true
}

func (e Expedition) Conditions(m monster.Monster) []Condition {
	return []Condition{
		{Kind: CondLife, Need: e.Life, Have: m.Stats[monster.Life]},
		{Kind: CondFame, Need: e.Fame, Have: m.Fame},
		{Kind: CondGrade, Need: e.Rank, Have: m.Rank},
	}
}

func RankLetter(rank int) string {
	const letters = "EDCBAS"
	if rank < 0 || rank >= len(letters) {
		return "?"
	}
	return letters[rank : rank+1]
}

func StartEnergy(life int) int { return min(max(life/70*70, 140), 700) }

func BreakChance(pow int) float64 { return Expedition{}.BreakChance(pow) }

func FindChance(intelligence int) float64 { return min(0.05+0.10*float64(intelligence/100), 0.95) }
