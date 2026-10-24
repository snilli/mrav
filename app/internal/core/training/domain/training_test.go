package domain_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/training/domain"
)

func everyGain() domain.Gains {
	var list []domain.Gain
	for stage := range 10 {
		for rate := range 5 {
			for _, kind := range []int{domain.KindLight, domain.KindHeavy, domain.KindHeavySecondary} {
				list = append(list, domain.Gain{
					LifeStage: stage, GrowthRate: rate, Kind: kind, Text: fmt.Sprintf("s%d-g%d-k%d", stage, rate, kind),
				})
			}
		}
	}
	return domain.NewGains(list)
}

var _ = Describe("Gains", func() {
	It("gives the text of a gain, and N/A where the table has none", func() {
		g := domain.NewGains([]domain.Gain{{LifeStage: 3, GrowthRate: 2, Kind: domain.KindHeavy, Text: "3-6 (3/4/5/6)"}})
		Expect(g.Text(3, 2, domain.KindHeavy)).To(Equal("3-6 (3/4/5/6)"))
		Expect(g.Text(3, 2, domain.KindLight)).To(Equal("N/A"))
		Expect(g.Text(10, 2, domain.KindHeavy)).To(Equal("N/A"))
	})
})

var _ = Describe("Drills", func() {
	It("lists the drills in the order of the page, a heavy one with its two stats, as a copy that is safe to change", func() {
		drills := domain.Drills()
		Expect(drills).To(HaveLen(10))
		Expect(drills[0]).To(Equal(domain.Drill{Name: "Domino", Primary: monster.Pow}))
		Expect(drills[6]).To(Equal(domain.Drill{Name: "Pull", Heavy: true, Primary: monster.Pow, Secondary: monster.Life, Slot: 6}))

		drills[0].Name = "Changed"
		Expect(domain.Drills()[0].Name).To(Equal("Domino"))
	})
})

var _ = Describe("Build", func() {
	m := monster.Monster{LifeStage: 4, Growth: [monster.StatCount]int{1, 2, 3, 4, 0, 2}}

	It("offers the ten drills of the page, in its order, with one swim", func() {
		var names []string
		for _, d := range domain.Build(m, everyGain()).Drills {
			names = append(names, d.Drill.Name)
		}
		Expect(names).To(Equal([]string{
			"Domino", "Study", "Run", "Shoot", "Dodge", "Endure", "Pull", "Meditate", "Leap",
			"Swim",
		}))
	})

	DescribeTable("reads each drill's gains from the growth rate of the stat it trains",
		func(name, primary, secondary string) {
			for _, d := range domain.Build(m, everyGain()).Drills {
				if d.Drill.Name == name {
					Expect(d.Primary).To(Equal(primary))
					Expect(d.Secondary).To(Equal(secondary))
					return
				}
			}
			Fail("no drill called " + name)
		},
		Entry("Domino trains Pow", "Domino", "s4-g2-k0", "N/A"),
		Entry("Study trains Int", "Study", "s4-g3-k0", "N/A"),
		Entry("Run trains Life", "Run", "s4-g1-k0", "N/A"),
		Entry("Shoot trains Skill", "Shoot", "s4-g4-k0", "N/A"),
		Entry("Dodge trains Speed", "Dodge", "s4-g0-k0", "N/A"),
		Entry("Endure trains Defense", "Endure", "s4-g2-k0", "N/A"),
		Entry("Pull is heavy: Pow a lot and Life a little", "Pull", "s4-g2-k1", "s4-g1-k3"),
		Entry("Meditate: Int and Skill", "Meditate", "s4-g3-k1", "s4-g4-k3"),
		Entry("Leap: Speed and Int", "Leap", "s4-g0-k1", "s4-g3-k3"),
		Entry("Swim: Defense a lot and Life a little, as the games' own results show", "Swim", "s4-g2-k1", "s4-g1-k3"),
	)

	It("reads the motivation of each drill from the monster", func() {
		s := m
		s.Motivation = [monster.DrillCount]int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
		var got []int
		for _, d := range domain.Build(s, everyGain()).Drills {
			got = append(got, d.Motivation)
		}
		Expect(got).To(Equal([]int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}))
	})

	It("says which drills the monster specialises in, by the bit of each, the swim by the pool", func() {
		s := m
		s.TrainBoost = 1<<0 | 1<<6 | 1<<9
		var got []string
		for _, d := range domain.Build(s, everyGain()).Drills {
			if d.Specialty {
				got = append(got, d.Drill.Name)
			}
		}
		Expect(got).To(Equal([]string{"Domino", "Pull", "Swim"}))
	})

	It("recommends a drill the monster specialises in and is keen on, keen being the green of MR2AV's motivation boxes", func() {
		s := m
		s.TrainBoost = 1 << 0
		s.Motivation[0] = 81
		Expect(domain.Build(s, everyGain()).Drills[0].Recommended).To(BeTrue())

		s.Motivation[0] = 80
		Expect(domain.Build(s, everyGain()).Drills[0].Recommended).To(BeFalse())

		s.Motivation[0], s.TrainBoost = 100, 0
		Expect(domain.Build(s, everyGain()).Drills[0].Recommended).To(BeFalse())
	})

	It("has nothing for a growth rate past A, or a life stage past the tenth", func() {
		s := m
		s.Growth[monster.Pow] = 5
		Expect(domain.Build(s, everyGain()).Drills[0].Primary).To(Equal("N/A"))

		s = m
		s.LifeStage = 10
		Expect(domain.Build(s, everyGain()).Drills[0].Primary).To(Equal("N/A"))
	})

	It("says which drills and places the monster specialises in, by the bits of its training boost", func() {
		s := m
		s.TrainBoost = 1<<0 | 1<<2 | 1<<15
		got := domain.Build(s, everyGain()).Specialties
		for bit, on := range got {
			Expect(on).To(Equal(bit == 0 || bit == 2 || bit == 15), domain.SpecialtyNames[bit])
		}
	})

	It("names the specialties as MR2AV does", func() {
		Expect(domain.SpecialtyNames).To(Equal([domain.SpecialtyCount]string{
			"Domino", "Study", "Run", "Shoot", "Dodge", "Endure", "Pull", "Meditate", "Leap",
			"Pool", "Torble Sea", "Papas Mountain", "Mandy Desert", "Parepare Jungle", "Kawrea Volcano", "???",
		}))
	})
})
