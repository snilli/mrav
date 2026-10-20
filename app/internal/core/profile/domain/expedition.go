package domain

import (
	monster "app/internal/core/monster/domain"
)

// The cost of what a monster does on an expedition, in the energy it begins with, and the energy below which the
// expedition fails and costs the monster weeks of its life.
const (
	MoveCost     = 1
	ObstacleCost = 5
	SearchCost   = 10
	FailBelow    = -9
	FailWeeks    = 12 // of life, lost when the expedition ends under FailBelow
)

// Expedition is a trip a monster is invited on, in the same month and week of every year, when it is fit enough
// and on the ranch as the week turns. The dates and what they ask of the monster are the game's own rules, as
// players' guides keep them.
type Expedition struct {
	Name        string
	Host        string // who invites the monster, where the guide says
	Month, Week int
	Life, Fame  int // the least a monster needs of each
	Rank        int // the lowest grade of the monster, from 0 for E to 5 for S
}

// Expeditions are the trips, in the order of the calendar. The first expedition, Phoenix Volcano Adventure, is not one of
// them: it can be made once only, and the game does not say, where this viewer reads it, whether it was.
var Expeditions = []Expedition{
	{Name: "Torles Mountains", Host: "Rovest", Month: 2, Week: 1, Life: 210, Fame: 50, Rank: 2},
	{Name: "Kawrea Volcano", Host: "Dr. Talico", Month: 6, Week: 1, Life: 280, Fame: 50, Rank: 3},
	{Name: "Parepare Jungle", Host: "Kavaro", Month: 10, Week: 1, Life: 140, Fame: 40, Rank: 2},
}

// ConditionKind is what an invitation asks of a monster.
type ConditionKind int

const (
	CondLife ConditionKind = iota
	CondFame
	CondGrade
)

// Condition is one thing an invitation asks of the monster: the least it must have, and what it has. A grade is its rank,
// from 0 for E to 5 for S.
type Condition struct {
	Kind       ConditionKind
	Need, Have int
}

// Met says whether the monster has enough.
func (c Condition) Met() bool { return c.Have >= c.Need }

// Fit says whether the monster has enough for every condition of the invitation.
func (e Expedition) Fit(m monster.Monster) bool {
	for _, c := range e.Conditions(m) {
		if !c.Met() {
			return false
		}
	}
	return true
}

// Conditions are what the invitation asks of a monster: its life, its fame and its grade.
func (e Expedition) Conditions(m monster.Monster) []Condition {
	return []Condition{
		{Kind: CondLife, Need: e.Life, Have: m.Stats[monster.Life]},
		{Kind: CondFame, Need: e.Fame, Have: m.Fame},
		{Kind: CondGrade, Need: e.Rank, Have: m.Rank},
	}
}

// RankLetter is a monster's rank as the game writes it, E to S, or ? for a number that is no rank.
func RankLetter(rank int) string {
	const letters = "EDCBAS"
	if rank < 0 || rank >= len(letters) {
		return "?"
	}
	return letters[rank : rank+1]
}

// StartEnergy is the energy a monster begins an expedition with: 70 for every 70 of its life, at least 140 and at
// most 700.
func StartEnergy(life int) int { return min(max(life/70*70, 140), 700) }

// BreakChance is the chance that an attempt on an obstacle breaks it, for a monster with this much POW. The guide
// gives 50% at 350 rising to 100% at 500 for the maps of the trips, and says only that it climbs steeply toward 300.
//
// shortcut: the line through those two points, carried below 350 and kept above 5%, where the guide gives no
// numbers; replace it when the game's own table is known.
func BreakChance(pow int) float64 { return min(max(0.5+float64(pow-350)/300, 0.05), 1) }

// FindChance is the chance that a search finds the way through a gate, for a monster with this much intelligence: 5%
// below 100, 10% more for each 100, and 95% from 900.
func FindChance(intelligence int) float64 { return min(0.05+0.10*float64(intelligence/100), 0.95) }
