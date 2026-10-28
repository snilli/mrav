package domain_test

import (
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/drill/domain"
	monster "app/internal/core/monster/domain"
)

var (
	results = []domain.Result{domain.Cheat, domain.Success, domain.Great}
	roles   = []domain.Role{domain.Light, domain.Primary, domain.Secondary, domain.Drop}
)

func code(result domain.Result, role domain.Role, bonus bool, stage, rate int) int {
	c := slices.Index(results, result)*10000 + slices.Index(roles, role)*1000 + stage*10 + rate
	if bonus {
		c += 100
	}
	return c
}

func everyCell() domain.Gains {
	var list []domain.Cell
	for _, result := range results {
		for _, role := range roles {
			for _, bonus := range []bool{false, true} {
				if role == domain.Drop && bonus {
					continue
				}
				for stage := range 10 {
					for rate := range 5 {
						list = append(list, domain.Cell{
							Result: result, Role: role, Bonus: bonus, LifeStage: stage, GrowthRate: rate,
							Slots: [][]int{{code(result, role, bonus, stage, rate)}},
						})
					}
				}
			}
		}
	}
	return domain.NewGains(list)
}

func effect(stat monster.Stat, value int) domain.Effect {
	return domain.Effect{Stat: stat, Low: value, High: value}
}

var _ = Describe("Gains", func() {
	It("gives the choices of a cell, and nothing where the table has none", func() {
		g := domain.NewGains([]domain.Cell{{Result: domain.Great, Role: domain.Primary, Bonus: true, LifeStage: 3, GrowthRate: 2, Slots: [][]int{{4, 8}, {6, 10}}}})
		Expect(g.Slots(domain.Great, domain.Primary, true, 3, 2)).To(Equal([][]int{{4, 8}, {6, 10}}))
		Expect(g.Slots(domain.Great, domain.Primary, false, 3, 2)).To(BeEmpty())
		Expect(g.Slots(domain.Success, domain.Primary, true, 3, 2)).To(BeEmpty())
		Expect(g.Slots(domain.Great, domain.Primary, true, 10, 2)).To(BeEmpty())
	})
})

var _ = Describe("Drills", func() {
	It("lists the drills in the order of the page, a heavy one with the stat it drops, as a copy that is safe to change", func() {
		drills := domain.Drills()
		Expect(drills).To(HaveLen(10))
		Expect(drills[0]).To(Equal(domain.Drill{Name: "Domino", Primary: monster.Pow}))
		Expect(drills[6]).To(Equal(domain.Drill{Name: "Pull", Heavy: true, Primary: monster.Pow, Secondary: monster.Life, Drop: monster.Speed, Slot: 6}))
		Expect(drills[9]).To(Equal(domain.Drill{Name: "Swim", Heavy: true, Primary: monster.Defense, Secondary: monster.Life, Drop: monster.Int, Slot: 9}))

		drills[0].Name = "Changed"
		Expect(domain.Drills()[0].Name).To(Equal("Domino"))
	})
})

var _ = Describe("Build", func() {
	m := monster.Monster{LifeStage: 4, Growth: [monster.StatCount]int{1, 2, 3, 4, 0, 2}}

	success := func(t domain.Training, name string) []domain.Effect {
		for _, d := range t.Drills {
			if d.Drill.Name == name {
				return d.Outcomes[1].Effects
			}
		}
		Fail("no drill called " + name)
		return nil
	}

	It("offers the ten drills of the page, in its order, with one swim", func() {
		var names []string
		for _, d := range domain.Build(m, everyCell()).Drills {
			names = append(names, d.Drill.Name)
		}
		Expect(names).To(Equal([]string{"Domino", "Study", "Run", "Shoot", "Dodge", "Endure", "Pull", "Meditate", "Leap", "Swim"}))
	})

	It("gives every drill the three results of the page, Cheat, Success and Great, in that order", func() {
		for _, d := range domain.Build(m, everyCell()).Drills {
			var got []domain.Result
			for _, o := range d.Outcomes {
				got = append(got, o.Result)
			}
			Expect(got).To(Equal([]domain.Result{domain.Cheat, domain.Success, domain.Great}), d.Drill.Name)
		}
	})

	DescribeTable("reads each drill's numbers from the growth rate of the stat it trains, the drop from the stat it drops",
		func(name string, want ...domain.Effect) {
			Expect(success(domain.Build(m, everyCell()), name)).To(Equal(want))
		},
		Entry("Domino trains Pow", "Domino", effect(monster.Pow, code(domain.Success, domain.Light, false, 4, 2))),
		Entry("Study trains Int", "Study", effect(monster.Int, code(domain.Success, domain.Light, false, 4, 3))),
		Entry("Run trains Life", "Run", effect(monster.Life, code(domain.Success, domain.Light, false, 4, 1))),
		Entry("Shoot trains Skill", "Shoot", effect(monster.Skill, code(domain.Success, domain.Light, false, 4, 4))),
		Entry("Dodge trains Speed", "Dodge", effect(monster.Speed, code(domain.Success, domain.Light, false, 4, 0))),
		Entry("Endure trains Defense", "Endure", effect(monster.Defense, code(domain.Success, domain.Light, false, 4, 2))),
		Entry("Pull: Pow a lot, Life a little, Speed down", "Pull",
			effect(monster.Pow, code(domain.Success, domain.Primary, false, 4, 2)),
			effect(monster.Life, code(domain.Success, domain.Secondary, false, 4, 1)),
			effect(monster.Speed, code(domain.Success, domain.Drop, false, 4, 0))),
		Entry("Meditate: Int, Skill, Defense down", "Meditate",
			effect(monster.Int, code(domain.Success, domain.Primary, false, 4, 3)),
			effect(monster.Skill, code(domain.Success, domain.Secondary, false, 4, 4)),
			effect(monster.Defense, code(domain.Success, domain.Drop, false, 4, 2))),
		Entry("Leap: Speed, Int, Pow down", "Leap",
			effect(monster.Speed, code(domain.Success, domain.Primary, false, 4, 0)),
			effect(monster.Int, code(domain.Success, domain.Secondary, false, 4, 3)),
			effect(monster.Pow, code(domain.Success, domain.Drop, false, 4, 2))),
		Entry("Swim: Defense a lot and Life a little, as the games' own results show, Int down", "Swim",
			effect(monster.Defense, code(domain.Success, domain.Primary, false, 4, 2)),
			effect(monster.Life, code(domain.Success, domain.Secondary, false, 4, 1)),
			effect(monster.Int, code(domain.Success, domain.Drop, false, 4, 3))),
	)

	It("reads the other two results from their own tables", func() {
		d := domain.Build(m, everyCell()).Drills[0]
		Expect(d.Outcomes[0].Effects).To(Equal([]domain.Effect{effect(monster.Pow, code(domain.Cheat, domain.Light, false, 4, 2))}))
		Expect(d.Outcomes[2].Effects).To(Equal([]domain.Effect{effect(monster.Pow, code(domain.Great, domain.Light, false, 4, 2))}))
	})

	It("reads the tables with the bonus for a drill the monster specialises in, and the one table of the drop, for the page has none with a bonus", func() {
		s := m
		s.TrainBoost = 1<<6 | 1<<0
		t := domain.Build(s, everyCell())
		Expect(success(t, "Pull")).To(Equal([]domain.Effect{
			effect(monster.Pow, code(domain.Success, domain.Primary, true, 4, 2)),
			effect(monster.Life, code(domain.Success, domain.Secondary, true, 4, 1)),
			effect(monster.Speed, code(domain.Success, domain.Drop, false, 4, 0)),
		}))
		Expect(success(t, "Domino")).To(Equal([]domain.Effect{effect(monster.Pow, code(domain.Success, domain.Light, true, 4, 2))}))
		Expect(success(t, "Study")).To(Equal([]domain.Effect{effect(monster.Int, code(domain.Success, domain.Light, false, 4, 3))}))
	})

	It("takes the lowest and the highest of all the values of a cell as its range, a pair counting both of its values, and a drop below zero", func() {
		g := domain.NewGains([]domain.Cell{
			{Result: domain.Great, Role: domain.Primary, LifeStage: 4, GrowthRate: 2, Slots: [][]int{{4, 8}, {6, 10}}},
			{Result: domain.Cheat, Role: domain.Drop, LifeStage: 4, GrowthRate: 0, Slots: [][]int{{-3}, {-4}, {-1}, {-2}}},
		})
		pull := domain.Build(m, g).Drills[6]
		Expect(pull.Outcomes[2].Effects).To(Equal([]domain.Effect{{Stat: monster.Pow, Low: 4, High: 10}}))
		Expect(pull.Outcomes[0].Effects).To(Equal([]domain.Effect{{Stat: monster.Speed, Low: -4, High: -1}}))
		Expect(pull.Outcomes[1].Effects).To(BeEmpty())
	})

	It("reads the motivation of each drill from the monster", func() {
		s := m
		s.Motivation = [monster.DrillCount]int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
		var got []int
		for _, d := range domain.Build(s, everyCell()).Drills {
			got = append(got, d.Motivation)
		}
		Expect(got).To(Equal([]int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}))
	})

	It("says which drills the monster specialises in, by the bit of each, the swim by the pool", func() {
		s := m
		s.TrainBoost = 1<<0 | 1<<6 | 1<<9
		var got []string
		for _, d := range domain.Build(s, everyCell()).Drills {
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
		Expect(domain.Build(s, everyCell()).Drills[0].Recommended).To(BeTrue())

		s.Motivation[0] = 80
		Expect(domain.Build(s, everyCell()).Drills[0].Recommended).To(BeFalse())

		s.Motivation[0], s.TrainBoost = 100, 0
		Expect(domain.Build(s, everyCell()).Drills[0].Recommended).To(BeFalse())
	})

	It("has no effects for a growth rate past A, or a life stage past the tenth", func() {
		s := m
		s.Growth[monster.Pow] = 5
		Expect(domain.Build(s, everyCell()).Drills[0].Outcomes[1].Effects).To(BeEmpty())

		s = m
		s.LifeStage = 10
		Expect(domain.Build(s, everyCell()).Drills[0].Outcomes[1].Effects).To(BeEmpty())
	})

	It("says which drills and places the monster specialises in, by the bits of its training boost", func() {
		s := m
		s.TrainBoost = 1<<0 | 1<<2 | 1<<15
		got := domain.Build(s, everyCell()).Specialties
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
