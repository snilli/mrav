package domain

import (
	"slices"

	monster "app/internal/core/monster/domain"
)

type Result string

type Role string

const (
	Cheat   Result = "cheat"
	Success Result = "success"
	Great   Result = "great"

	Light     Role = "light"
	Primary   Role = "primary"
	Secondary Role = "secondary"
	Drop      Role = "drop"

	SpecialtyCount = 16

	recommendedMotivation = 81
)

var SpecialtyNames = [SpecialtyCount]string{
	"Domino", "Study", "Run", "Shoot", "Dodge", "Endure", "Pull", "Meditate", "Leap",
	"Pool", "Torble Sea", "Papas Mountain", "Mandy Desert", "Parepare Jungle", "Kawrea Volcano", "???",
}

var resultOrder = [...]Result{Cheat, Success, Great}

type Drill struct {
	Name      string
	Heavy     bool
	Primary   monster.Stat
	Secondary monster.Stat
	Drop      monster.Stat
	Slot      int
}

var drills = [...]Drill{
	{Name: "Domino", Primary: monster.Pow, Slot: 0},
	{Name: "Study", Primary: monster.Int, Slot: 1},
	{Name: "Run", Primary: monster.Life, Slot: 2},
	{Name: "Shoot", Primary: monster.Skill, Slot: 3},
	{Name: "Dodge", Primary: monster.Speed, Slot: 4},
	{Name: "Endure", Primary: monster.Defense, Slot: 5},
	{Name: "Pull", Heavy: true, Primary: monster.Pow, Secondary: monster.Life, Drop: monster.Speed, Slot: 6},
	{Name: "Meditate", Heavy: true, Primary: monster.Int, Secondary: monster.Skill, Drop: monster.Defense, Slot: 7},
	{Name: "Leap", Heavy: true, Primary: monster.Speed, Secondary: monster.Int, Drop: monster.Pow, Slot: 8},
	{Name: "Swim", Heavy: true, Primary: monster.Defense, Secondary: monster.Life, Drop: monster.Int, Slot: 9},
}

func Drills() []Drill { return slices.Clone(drills[:]) }

type Cell struct {
	Result     Result
	Role       Role
	Bonus      bool
	LifeStage  int
	GrowthRate int
	Slots      [][]int
}

type cellKey struct {
	result Result
	role   Role
	bonus  bool
	stage  int
	rate   int
}

type Gains struct{ byKey map[cellKey][][]int }

func NewGains(cells []Cell) Gains {
	g := Gains{byKey: make(map[cellKey][][]int, len(cells))}
	for _, c := range cells {
		g.byKey[cellKey{c.Result, c.Role, c.Bonus, c.LifeStage, c.GrowthRate}] = c.Slots
	}
	return g
}

func (g Gains) Slots(result Result, role Role, bonus bool, stage, rate int) [][]int {
	return g.byKey[cellKey{result, role, bonus, stage, rate}]
}

type Effect struct {
	Stat monster.Stat
	Low  int
	High int
}

type Outcome struct {
	Result  Result
	Effects []Effect
}

type DrillGains struct {
	Drill       Drill
	Outcomes    []Outcome
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
		g := DrillGains{Drill: d, Motivation: m.Motivation[d.Slot], Specialty: t.Specialties[d.Slot]}
		g.Recommended = g.Specialty && g.Motivation >= recommendedMotivation
		for _, result := range resultOrder {
			g.Outcomes = append(g.Outcomes, Outcome{Result: result, Effects: effects(m, d, result, g.Specialty, gains)})
		}
		t.Drills = append(t.Drills, g)
	}
	return t
}

func effects(m monster.Monster, d Drill, result Result, bonus bool, gains Gains) []Effect {
	var out []Effect
	add := func(stat monster.Stat, role Role, bonus bool) {
		if low, high, ok := span(gains.Slots(result, role, bonus, m.LifeStage, m.Growth[stat])); ok {
			out = append(out, Effect{Stat: stat, Low: low, High: high})
		}
	}
	if !d.Heavy {
		add(d.Primary, Light, bonus)
		return out
	}
	add(d.Primary, Primary, bonus)
	add(d.Secondary, Secondary, bonus)
	add(d.Drop, Drop, false)
	return out
}

func span(slots [][]int) (low, high int, ok bool) {
	for _, slot := range slots {
		for _, v := range slot {
			if !ok {
				low, high, ok = v, v, true
				continue
			}
			low, high = min(low, v), max(high, v)
		}
	}
	return low, high, ok
}
