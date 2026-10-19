// Package domain is the Training page: which drills and places a monster does well at, and what each drill
// gives it. Pure Go.
package domain

import (
	"slices"

	monster "app/internal/core/monster/domain"
)

const (
	// The kinds of gain: the stat a light drill trains, the main stat of a heavy drill, and its other stat.
	KindLight          = 0
	KindHeavy          = 1
	KindHeavySecondary = 3

	// SpecialtyCount is how many specialties the training boost can name, one to a bit.
	SpecialtyCount = 16

	noGain = "N/A"
)

// SpecialtyNames are what the bits of monster.Monster.TrainBoost stand for, the lowest bit first.
var SpecialtyNames = [SpecialtyCount]string{
	"Domino", "Study", "Run", "Shoot", "Dodge", "Endure", "Pull", "Meditate", "Leap",
	"Pool", "Torble Sea", "Papas Mountain", "Mandy Desert", "Parepare Jungle", "Kawrea Volcano", "???",
}

// Drill is a kind of training. A light drill trains one stat; a heavy drill trains its main stat a lot and
// another a little.
type Drill struct {
	Name      string
	Heavy     bool
	Primary   monster.Stat
	Secondary monster.Stat // only for a heavy drill
	// Slot is the drill's place among the monster's motivations and among the bits of its training boost. Both
	// swims are the pool's, so they share one.
	Slot int
}

// recommendedMotivation is the motivation from which a drill is keen: the 81 percent where MR2AV's boxes go green.
const recommendedMotivation = 81

// drills are the drills in the order the page lists them.
var drills = [...]Drill{
	{Name: "Domino", Primary: monster.Pow, Slot: 0},
	{Name: "Study", Primary: monster.Int, Slot: 1},
	{Name: "Run", Primary: monster.Life, Slot: 2},
	{Name: "Shoot", Primary: monster.Skill, Slot: 3},
	{Name: "Dodge", Primary: monster.Speed, Slot: 4},
	{Name: "Endure", Primary: monster.Defense, Slot: 5},
	{Name: "Pull", Heavy: true, Primary: monster.Pow, Secondary: monster.Life, Slot: 6},
	{Name: "Meditate", Heavy: true, Primary: monster.Int, Secondary: monster.Skill, Slot: 7},
	{Name: "Leap", Heavy: true, Primary: monster.Speed, Secondary: monster.Int, Slot: 8},
	{Name: "Swim (Vanilla)", Heavy: true, Primary: monster.Life, Secondary: monster.Defense, Slot: 9},
	{Name: "Swim (Patched/HM)", Heavy: true, Primary: monster.Defense, Secondary: monster.Life, Slot: 9},
}

// Drills are the drills of the page, in its order.
func Drills() []Drill { return slices.Clone(drills[:]) }

// Gain is what a drill gives, in words, for one kind of gain at a life stage and a growth rate.
type Gain struct {
	LifeStage  int
	GrowthRate int
	Kind       int
	Text       string
}

type gainKey struct{ stage, rate, kind int }

// Gains is the table of gains.
type Gains struct{ byKey map[gainKey]string }

func NewGains(list []Gain) Gains {
	g := Gains{byKey: make(map[gainKey]string, len(list))}
	for _, gain := range list {
		g.byKey[gainKey{gain.LifeStage, gain.GrowthRate, gain.Kind}] = gain.Text
	}
	return g
}

// Text is the gain at a life stage and growth rate, or N/A where the table has none, such as for a growth
// rate past A.
func (g Gains) Text(stage, rate, kind int) string {
	if text, ok := g.byKey[gainKey{stage, rate, kind}]; ok {
		return text
	}
	return noGain
}

// DrillGains is what one drill gives the monster: Primary for the stat it trains, Secondary for the other stat
// of a heavy drill, and N/A for a light drill's. With it go how keen the monster is for the drill, whether the
// drill is one it specialises in, and whether the two together make it worth doing.
type DrillGains struct {
	Drill       Drill
	Primary     string
	Secondary   string
	Motivation  int  // percent
	Specialty   bool // the monster is good at this drill
	Recommended bool // a specialty the monster is keen on
}

// Training is the Training page.
type Training struct {
	Specialties [SpecialtyCount]bool // by bit, as SpecialtyNames
	Drills      []DrillGains
}

// Build reads what each drill gives a monster from its life stage and the growth rates of the stats it trains.
func Build(m monster.Monster, gains Gains) Training {
	var t Training
	for bit := range t.Specialties {
		t.Specialties[bit] = m.TrainBoost&(1<<bit) != 0
	}
	for _, d := range drills {
		g := DrillGains{Drill: d, Secondary: noGain, Motivation: m.Motivation[d.Slot], Specialty: t.Specialties[d.Slot]}
		g.Recommended = g.Specialty && g.Motivation >= recommendedMotivation
		if d.Heavy {
			g.Primary = gains.Text(m.LifeStage, m.Growth[d.Primary], KindHeavy)
			g.Secondary = gains.Text(m.LifeStage, m.Growth[d.Secondary], KindHeavySecondary)
		} else {
			g.Primary = gains.Text(m.LifeStage, m.Growth[d.Primary], KindLight)
		}
		t.Drills = append(t.Drills, g)
	}
	return t
}
