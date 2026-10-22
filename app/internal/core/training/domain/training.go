package domain

import (
	"slices"

	monster "app/internal/core/monster/domain"
)

const (
	KindLight          = 0
	KindHeavy          = 1
	KindHeavySecondary = 3

	SpecialtyCount = 16

	noGain = "N/A"
)

var SpecialtyNames = [SpecialtyCount]string{
	"Domino", "Study", "Run", "Shoot", "Dodge", "Endure", "Pull", "Meditate", "Leap",
	"Pool", "Torble Sea", "Papas Mountain", "Mandy Desert", "Parepare Jungle", "Kawrea Volcano", "???",
}

type Drill struct {
	Name      string
	Heavy     bool
	Primary   monster.Stat
	Secondary monster.Stat
	Slot      int
}

const recommendedMotivation = 81

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
	{Name: "Swim", Heavy: true, Primary: monster.Life, Secondary: monster.Defense, Slot: swimSlot},
}

const swimSlot = 9

var patchedSwim = Drill{Name: "Swim", Heavy: true, Primary: monster.Defense, Secondary: monster.Life, Slot: swimSlot}

func Drills() []Drill { return slices.Clone(drills[:]) }

type Gain struct {
	LifeStage  int
	GrowthRate int
	Kind       int
	Text       string
}

type gainKey struct{ stage, rate, kind int }

type Gains struct{ byKey map[gainKey]string }

func NewGains(list []Gain) Gains {
	g := Gains{byKey: make(map[gainKey]string, len(list))}
	for _, gain := range list {
		g.byKey[gainKey{gain.LifeStage, gain.GrowthRate, gain.Kind}] = gain.Text
	}
	return g
}

func (g Gains) Text(stage, rate, kind int) string {
	if text, ok := g.byKey[gainKey{stage, rate, kind}]; ok {
		return text
	}
	return noGain
}

type DrillGains struct {
	Drill       Drill
	Primary     string
	Secondary   string
	Motivation  int
	Specialty   bool
	Recommended bool
}

type Training struct {
	Specialties [SpecialtyCount]bool
	Drills      []DrillGains
}

func Build(m monster.Monster, gains Gains) Training {
	var t Training
	for bit := range t.Specialties {
		t.Specialties[bit] = m.TrainBoost&(1<<bit) != 0
	}
	for _, d := range drills {
		if d.Slot == swimSlot && m.PatchedSwim {
			d = patchedSwim
		}
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
