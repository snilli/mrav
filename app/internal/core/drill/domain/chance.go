package domain

import (
	"math"

	monster "app/internal/core/monster/domain"
)

const rolls = 256

var levelEase = [...]int{1: 75, 2: 125, 3: 200}

type Chances struct {
	Success float64
	Failure float64
	Great   float64
	Cheat   float64
	Truancy float64
}

func (c Chances) Of(r Result) float64 {
	switch r {
	case Cheat:
		return c.Cheat
	case Great:
		return c.Great
	}
	return c.Success
}

func (d Drill) Level(m monster.Monster) int {
	if d.Heavy {
		v := m.Stats[d.Primary] + m.Stats[d.Secondary]/2
		switch {
		case v > 1050:
			return 3
		case v >= 450:
			return 2
		}
		return 1
	}
	switch v := m.Stats[d.Primary]; {
	case v >= 700:
		return 3
	case v >= 300:
		return 2
	}
	return 1
}

func ChancesOf(m monster.Monster, d Drill) Chances {
	nature := max(-100, min(100, m.Nature))
	motivation := m.Motivation[d.Slot]

	truancy := truancyOf(m.Fatigue, nature)
	cheat := cheatOf(nature, motivation) * (1 - truancy)
	failure := failureOf(m, d) * (1 - cheat - truancy)
	great := greatOf(nature, motivation) * (1 - failure - cheat - truancy)

	c := Chances{Failure: percent(failure), Great: percent(great), Cheat: percent(cheat), Truancy: percent(truancy)}
	c.Success = math.Round(float64(100-c.Failure-c.Great-c.Cheat-c.Truancy)*10) / 10
	return c
}

func percent(p float64) float64 { return math.Round(float64(p*1000)) / 10 }

func floorDiv(a, b int) int { return int(math.Floor(float64(a) / float64(b))) }

func count(hit func(i int) bool) float64 {
	n := 0
	for i := range rolls {
		if hit(i) {
			n++
		}
	}
	return float64(n) / rolls
}

func truancyOf(fatigue, nature int) float64 {
	if (nature >= 0 && fatigue < 80) || (nature < 0 && fatigue < 50) {
		return 0
	}
	nat := floorDiv(nature, 2)
	return count(func(i int) bool { return fatigue >= 150-i%(100-nat) })
}

func cheatOf(nature, motivation int) float64 {
	step, nat := 15, floorDiv(nature, 10)
	if nature < 0 {
		step, nat = 10, floorDiv(nature, 5)
	}
	mot := step * ((motivation + step - 1) / step)
	return count(func(i int) bool { return mot < 150-step*(nat+(29921875*i/25500000)%100) })
}

func failureOf(m monster.Monster, d Drill) float64 {
	hard := d.Hardness
	if !d.Heavy {
		hard = 20 * (m.LifeStage + 1)
		if d.Primary != monster.Defense {
			hard += 10 * m.Growth[d.Primary+1]
		}
	}
	ease := levelEase[d.Level(m)]
	return count(func(i int) bool { return m.Fatigue > -ease+m.Fear+hard+5*(i%30) })
}

func greatOf(nature, motivation int) float64 {
	nat := floorDiv(nature, 10)
	if nature < 0 {
		nat = floorDiv(nature, 20)
	}
	return count(func(i int) bool { return motivation/10 > i%100-nat })
}
