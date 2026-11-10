package domain_test

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/drill/domain"
	monster "app/internal/core/monster/domain"
)

type chanceCase struct {
	Drill      string  `json:"drill"`
	Fatigue    int     `json:"fatigue"`
	Fear       int     `json:"fear"`
	LifeStage  int     `json:"life_stage"`
	Nature     int     `json:"nature"`
	Level      int     `json:"level"`
	Motivation int     `json:"motivation"`
	Nag        int     `json:"nag"`
	Success    float64 `json:"success"`
	Failure    float64 `json:"failure"`
	Great      float64 `json:"great"`
	Cheat      float64 `json:"cheat"`
	Truant     float64 `json:"truant"`
}

func drillNamed(name string) domain.Drill {
	for _, d := range domain.Drills() {
		if d.Name == name {
			return d
		}
	}
	Fail("no drill called " + name)
	return domain.Drill{}
}

func (c chanceCase) monster(d domain.Drill) monster.Monster {
	m := monster.Monster{Fatigue: c.Fatigue, Fear: c.Fear, LifeStage: c.LifeStage, Nature: c.Nature}
	m.Motivation[d.Slot] = c.Motivation
	if !d.Heavy && d.Primary != monster.Defense {
		m.Growth[d.Primary+1] = c.Nag
	}
	m.DrillLevel[d.Slot] = c.Level
	return m
}

var _ = Describe("the chances of a drill, worked out the way the Drill Calculator of LegendCup works them out", func() {
	var cases []chanceCase
	raw, err := os.ReadFile("testdata/chances.json")
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(raw, &cases)).To(Succeed())

	It("has the cases that the page worked out, for every drill, the edges of truancy and cheating, and random monsters", func() {
		Expect(len(cases)).To(BeNumerically(">=", 1000))
	})

	It("gives the chances that the page gives, for each case", func() {
		for _, c := range cases {
			d := drillNamed(c.Drill)
			m := c.monster(d)
			Expect(d.Level(m)).To(Equal(c.Level), "%+v", c)
			Expect(domain.ChancesOf(m, d)).To(Equal(domain.Chances{
				Success: c.Success, Failure: c.Failure, Great: c.Great, Cheat: c.Cheat, Truancy: c.Truant,
			}), "%+v", c)
		}
	})

	It("takes the stars of each drill from the game, by its slot", func() {
		var m monster.Monster
		m.DrillLevel[0], m.DrillLevel[6] = 3, 2
		Expect([]int{drillNamed("Domino").Level(m), drillNamed("Pull").Level(m)}).To(Equal([]int{3, 2}))
	})

	It("works the chances out for each drill of the monster", func() {
		m := monster.Monster{Fatigue: 30, Nature: 20}
		m.Motivation[6] = 90
		t := domain.Build(m, everyCell())
		Expect(t.Drills[6].Chances).To(Equal(domain.ChancesOf(m, t.Drills[6].Drill)))
	})
})
