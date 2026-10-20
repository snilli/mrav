package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/profile/domain"
)

var _ = Describe("what a search of a place turns up", func() {
	// mine is searched by intelligence: two tiers, from 0 and from 100, with a table of its own for a good nature.
	mine := func() domain.Spot {
		return domain.Spot{Name: "A", Chances: []domain.Chance{
			{MinStat: 0, Slot: 0, Item: "Gem", Pct: 10, Price: 1000},
			{MinStat: 0, Slot: 1, Item: "Nuts", Pct: 90, Price: 100},
			{MinStat: 100, Slot: 0, Item: "Gem", Pct: 20, Price: 1000},
			{MinStat: 100, Slot: 1, Item: "Nuts", Pct: 80, Price: 100},
			{Good: true, MinStat: 0, Slot: 0, Item: "Gem", Pct: 50, Price: 1000},
			{Good: true, MinStat: 100, Slot: 0, Item: "Gem", Pct: 60, Price: 1000},
		}}
	}
	// tree is searched by skill: nothing that sells in the first tier, a gem for sure in the second.
	tree := func() domain.Spot {
		return domain.Spot{Name: "B", BySkill: true, Chances: []domain.Chance{
			{MinStat: 0, Item: "Glue", Pct: 100, Price: 0},
			{MinStat: 200, Item: "Gem", Pct: 100, Price: 1000},
		}}
	}

	DescribeTable("is worth what the items at the chances of the tier sell for, on average",
		func(spot func() domain.Spot, by domain.Searcher, want float64) {
			value, _ := spot().Searched(by)
			Expect(value).To(BeNumerically("~", want, 1e-9))
		},
		Entry("the first tier of intelligence", mine, domain.Searcher{Intelligence: 99}, 190.0),
		Entry("the second, from where it starts", mine, domain.Searcher{Intelligence: 100}, 280.0),
		Entry("the second, as high as the stat goes", mine, domain.Searcher{Intelligence: 999}, 280.0),
		Entry("a good nature has a table of its own", mine, domain.Searcher{Intelligence: 150, GoodNature: true}, 600.0),
		Entry("a building searched by skill takes no notice of intelligence", tree, domain.Searcher{Intelligence: 999, Skill: 0}, 0.0),
		Entry("and goes by skill", tree, domain.Searcher{Intelligence: 0, Skill: 250}, 1000.0),
		Entry("an item that cannot be sold is worth nothing, whatever its chance", tree, domain.Searcher{Skill: 199}, 0.0),
		Entry("a stat below the first tier turns up nothing", func() domain.Spot {
			return domain.Spot{BySkill: true, Chances: []domain.Chance{{MinStat: 100, Item: "Gem", Pct: 100, Price: 1000}}}
		}, domain.Searcher{Skill: 99}, 0.0),
	)

	It("adds up the two rows that a table has for one item", func() {
		spot := mine()
		spot.Chances = append(spot.Chances, domain.Chance{MinStat: 0, Slot: 2, Item: "Gem", Pct: 10, Price: 1000})
		value, _ := spot.Searched(domain.Searcher{Intelligence: 0})
		Expect(value).To(BeNumerically("~", 100+100+90, 1e-9))
	})

	It("counts two items of one name for what each is, since the chances are of the items and not of the names", func() {
		spot := domain.Spot{Chances: []domain.Chance{
			{Slot: 0, Item: "Nuts", Pct: 50, Price: 100},
			{Slot: 1, Item: "Nuts", Pct: 50, Price: 300},
		}}
		value, _ := spot.Searched(domain.Searcher{})
		Expect(value).To(BeNumerically("~", 200, 1e-9))
	})

	It("gives a place that has no chances at all a worth of nothing and nothing that it turns up", func() {
		value, loot := domain.Spot{Name: "Z"}.Searched(domain.Searcher{Intelligence: 500})
		Expect(value).To(BeZero())
		Expect(loot).To(BeEmpty())
	})

	It("lists what it may turn up at the tier of the stat, the likeliest first and by the row of the table where it is as likely, and leaves out what the tier never turns up", func() {
		spot := domain.Spot{Chances: []domain.Chance{
			{MinStat: 0, Slot: 0, Item: "Old", Pct: 100, Price: 1},
			{MinStat: 100, Slot: 0, Item: "Mint Leaf", Pct: 12.5, Price: 100},
			{MinStat: 100, Slot: 1, Item: "Pixie Chips", Pct: 18.8, Price: 1000},
			{MinStat: 100, Slot: 2, Item: "Never", Pct: 0, Price: 5000},
			{MinStat: 100, Slot: 3, Item: "Golem Chips", Pct: 18.8, Price: 500},
			{MinStat: 100, Slot: 4, Item: "Griever", Pct: 12.5, Price: 50},
		}}
		_, loot := spot.Searched(domain.Searcher{Intelligence: 150})
		Expect(loot).To(Equal([]domain.Drop{
			{Item: "Pixie Chips", Pct: 18.8, Price: 1000},
			{Item: "Golem Chips", Pct: 18.8, Price: 500},
			{Item: "Mint Leaf", Pct: 12.5, Price: 100},
			{Item: "Griever", Pct: 12.5, Price: 50},
		}))
	})

	It("gives the same worth as the list that it turns up sells for at its chances", func() {
		value, loot := mine().Searched(domain.Searcher{Intelligence: 150})
		sum := 0.0
		for _, d := range loot {
			sum += d.Pct / 100 * float64(d.Price)
		}
		Expect(sum).To(BeNumerically("~", value, 1e-9))
	})

	Describe("For", func() {
		It("is the map with what a search of each place is worth and turns up for the searcher, and the map as it was left alone", func() {
			m := domain.ExpeditionMap{Name: "Cave", Spots: []domain.Spot{mine(), tree()}}
			got := m.For(domain.Searcher{Intelligence: 150, Skill: 250})

			Expect(got.Name).To(Equal("Cave"))
			Expect(got.Spots).To(HaveLen(2))
			Expect(got.Spots[0].Value).To(BeNumerically("~", 280, 1e-9))
			Expect(got.Spots[0].Loot).To(HaveLen(2))
			Expect(got.Spots[1].Value).To(BeNumerically("~", 1000, 1e-9))
			Expect(got.Spots[1].Chances).To(HaveLen(2), "the tables stay on the place, for another searcher")

			Expect(m.Spots[0].Value).To(BeZero(), "the map that was shared is not written to")
			Expect(m.Spots[0].Loot).To(BeNil())
		})

		It("keeps the places in their order, which the planner breaks ties by", func() {
			m := domain.ExpeditionMap{Spots: []domain.Spot{{Name: "B"}, {Name: "A"}, {Name: "C"}}}
			var names []string
			for _, s := range m.For(domain.Searcher{}).Spots {
				names = append(names, s.Name)
			}
			Expect(names).To(Equal([]string{"B", "A", "C"}))
		})
	})
})
