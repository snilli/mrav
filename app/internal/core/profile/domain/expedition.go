package domain

import (
	"fmt"
	"strconv"

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
	Life, Fame  int  // the least a monster needs of each
	Rank        int  // the lowest grade of the monster, from 0 for E to 5 for S
	First       bool // the first expedition, which is made once and unlocks all the others
}

// Expeditions are the trips, in the order of the calendar.
var Expeditions = []Expedition{
	{Name: "Torles Mountains", Host: "Rovest", Month: 2, Week: 1, Life: 210, Fame: 50, Rank: 2},
	{Name: "Phoenix Volcano Adventure", Month: 6, Week: 1, Life: 140, Fame: 40, Rank: 1, First: true},
	{Name: "Kawrea Volcano", Host: "Dr. Talico", Month: 6, Week: 1, Life: 280, Fame: 50, Rank: 3},
	{Name: "Parepare Jungle", Host: "Kavaro", Month: 10, Week: 1, Life: 140, Fame: 40, Rank: 2},
}

// Condition is one thing an invitation asks of the monster, and where the monster stands on it.
type Condition struct {
	Text string
	Have string
	Met  bool
}

// Conditions are what the invitation asks of a monster: its life, its fame and its grade.
func (e Expedition) Conditions(m monster.Monster) []Condition {
	life := m.Stats[monster.Life]
	return []Condition{
		{Text: fmt.Sprintf("LIF %d or more", e.Life), Have: strconv.Itoa(life), Met: life >= e.Life},
		{Text: fmt.Sprintf("Fame %d or more", e.Fame), Have: strconv.Itoa(m.Fame), Met: m.Fame >= e.Fame},
		{Text: fmt.Sprintf("Grade %s or higher", RankLetter(e.Rank)), Have: RankLetter(m.Rank), Met: m.Rank >= e.Rank},
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
