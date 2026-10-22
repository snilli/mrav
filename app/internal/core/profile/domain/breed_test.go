package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
)

type growth = [monster.StatCount]int

func mon(main, sub, guts int, g growth) monster.Monster {
	return monster.Monster{Genus: monster.Genus(main), SubGenus: monster.Genus(sub), GutsRate: guts, Growth: g}
}

var _ = Describe("ResolveBreed", func() {
	pixieG := growth{0, 1, 4, 3, 3, 0}
	pixie := domain.BreedRule{Main: 0, Sub: 0, Name: "Pixie", Otherwise: "Pixie (N/S)", Guts: new(7), Growth: &pixieG}

	It("names a breed whose growth rates are the standard ones, and says so when the guts rate is too", func() {
		b := domain.ResolveBreed([]domain.BreedRule{pixie}, nil, mon(0, 0, 7, pixieG))
		Expect(b).To(Equal(domain.Breed{Name: "Pixie", GutsOK: true}))
	})

	It("adds (N/S) to a name whose growth rates are not the standard ones", func() {
		b := domain.ResolveBreed([]domain.BreedRule{pixie}, nil, mon(0, 0, 7, growth{1, 1, 4, 3, 3, 0}))
		Expect(b.Name).To(Equal("Pixie (N/S)"))
	})

	It("keeps the name when only the guts rate is off, and flags the guts rate", func() {
		b := domain.ResolveBreed([]domain.BreedRule{pixie}, nil, mon(0, 0, 8, pixieG))
		Expect(b).To(Equal(domain.Breed{Name: "Pixie", GutsOK: false}))
	})

	It("takes the name a rule gives for growth rates that do not match, when it is not the name with (N/S)", func() {
		poison := domain.BreedRule{Main: 0, Sub: 38, Name: "Poison (DNA Capsule)", Otherwise: "Poison", Growth: &pixieG}
		Expect(domain.ResolveBreed([]domain.BreedRule{poison}, nil, mon(0, 38, 9, growth{4, 4, 4, 4, 4, 4})).Name).To(Equal("Poison"))
	})

	It("gives a fixed name to a breed that has no growth rates to check, and calls the guts rate right from 6 up", func() {
		kasumi := domain.BreedRule{Main: 0, Sub: 39, Name: "Kasumi"}
		Expect(domain.ResolveBreed([]domain.BreedRule{kasumi}, nil, mon(0, 39, 5, growth{}))).To(Equal(domain.Breed{Name: "Kasumi", GutsOK: false}))
		Expect(domain.ResolveBreed([]domain.BreedRule{kasumi}, nil, mon(0, 39, 6, growth{}))).To(Equal(domain.Breed{Name: "Kasumi", GutsOK: true}))
	})

	It("says undefined for a sub breed the main breed has no rule for, and no monster for a main breed with none", func() {
		Expect(domain.ResolveBreed([]domain.BreedRule{pixie}, nil, mon(0, 5, 7, pixieG)).Name).To(Equal("Undefined Monster"))
		Expect(domain.ResolveBreed(nil, nil, mon(0, 0, 7, pixieG)).Name).To(Equal("No Monster"))
	})

	Describe("a breed with several variants, chosen by guts rate and growth rates together", func() {
		g1, g2 := growth{2, 4, 4, 2, 1, 2}, growth{3, 3, 0, 1, 1, 3}
		moo := []domain.BreedRule{
			{Main: 1, Sub: 38, Name: "Moo (DNA Capsule)", Otherwise: "Moo (N/S)", Guts: new(19), MatchGuts: true, Growth: &g1},
			{Main: 1, Sub: 38, Name: "Moo (CD)", Otherwise: "Moo (N/S)", Guts: new(18), MatchGuts: true, Growth: &g2},
		}

		It("takes the first variant that matches", func() {
			Expect(domain.ResolveBreed(moo, nil, mon(1, 38, 19, g1)).Name).To(Equal("Moo (DNA Capsule)"))
			Expect(domain.ResolveBreed(moo, nil, mon(1, 38, 18, g2)).Name).To(Equal("Moo (CD)"))
		})

		It("does not match a variant on its growth rates alone", func() {
			Expect(domain.ResolveBreed(moo, nil, mon(1, 38, 18, g1)).Name).To(Equal("Moo (N/S)"))
		})

		It("calls the guts rate right when it is any variant's", func() {
			Expect(domain.ResolveBreed(moo, nil, mon(1, 38, 18, g2)).GutsOK).To(BeTrue())
			Expect(domain.ResolveBreed(moo, nil, mon(1, 38, 17, g2)).GutsOK).To(BeFalse())
		})
	})

	It("tells the first release of a breed from the patched one by its guts rate", func() {
		g := growth{0, 3, 2, 2, 4, 0}
		hare := []domain.BreedRule{
			{Main: 13, Sub: 0, Name: "Fairy Hare (MF2 v1.0)", Otherwise: "Fairy Hare (N/S)", Guts: new(125), MatchGuts: true, Growth: &g},
			{Main: 13, Sub: 0, Name: "Fairy Hare", Otherwise: "Fairy Hare (N/S)", Guts: new(12), Growth: &g},
		}
		Expect(domain.ResolveBreed(hare, nil, mon(13, 0, 125, g)).Name).To(Equal("Fairy Hare (MF2 v1.0)"))
		Expect(domain.ResolveBreed(hare, nil, mon(13, 0, 12, g)).Name).To(Equal("Fairy Hare"))
		Expect(domain.ResolveBreed(hare, nil, mon(13, 0, 12, growth{1, 3, 2, 2, 4, 0})).Name).To(Equal("Fairy Hare (N/S)"))
	})

	It("names a revived Henger and Gali the Protomessiah, whatever the rules say", func() {
		Expect(domain.ResolveBreed(nil, nil, mon(5, 15, 9, growth{})).Name).To(Equal("Protomessiah"))
	})

	Describe("a cocoon", func() {
		cocoonG := growth{2, 2, 3, 2, 1, 1}
		cocoons := []domain.CocoonRule{{Guts: 12, Growth: cocoonG, Name: "Cocooned Worm/Pixie"}}
		worm := domain.BreedRule{Main: 0, Sub: 36, Name: "Worm", Otherwise: "Worm (N/S)", Guts: new(15), Growth: &growth{4, 2, 2, 2, 0, 1}}

		It("is named by its guts rate and growth rates when it is the sub breed Worm of a breed that can cocoon", func() {
			b := domain.ResolveBreed([]domain.BreedRule{worm}, cocoons, mon(0, 36, 12, cocoonG))
			Expect(b).To(Equal(domain.Breed{Name: "Cocooned Worm/Pixie", GutsOK: true}))
		})

		It("is named from the ordinary rules when no cocoon matches", func() {
			b := domain.ResolveBreed([]domain.BreedRule{worm}, cocoons, mon(0, 36, 15, growth{4, 2, 2, 2, 0, 1}))
			Expect(b.Name).To(Equal("Worm"))
		})

		It("is not looked for in a breed that cannot cocoon", func() {
			other := domain.BreedRule{Main: 1, Sub: 36, Name: "Other", Otherwise: "Other (N/S)", Growth: &growth{}}
			Expect(domain.ResolveBreed([]domain.BreedRule{other}, cocoons, mon(1, 36, 12, cocoonG)).Name).To(Equal("Other (N/S)"))
		})
	})
})

var _ = Describe("PossibleCocoon", func() {
	DescribeTable("says which monsters might be cocooned",
		func(main, sub int, g growth, want bool) { Expect(domain.PossibleCocoon(main, sub, g)).To(Equal(want)) },
		Entry("a Beaclon of Beaclon whose growth rates are not a real one's", 4, 4, growth{1, 1, 1, 1, 1, 1}, true),
		Entry("a real Beaclon", 4, 4, growth{3, 3, 0, 1, 1, 3}, false),
		Entry("a Pixie with Worm for a sub breed", 0, 36, growth{}, true),
		Entry("a Naga with Worm for a sub breed", 37, 36, growth{}, true),
		Entry("a Dragon with Worm for a sub breed cannot", 1, 36, growth{}, false),
		Entry("a Pixie of something else", 0, 5, growth{}, false),
	)
})

var _ = Describe("cocoon readiness", func() {
	ready := monster.Monster{Genus: 36, Rank: 2, Age: 192, Spoil: 80, Fear: 80, Fatigue: 0, Stress: 29}

	It("shows for a Worm of rank C or lower, and not for anything else", func() {
		Expect(domain.ShowsCocoon(ready)).To(BeTrue())
		past := ready
		past.Rank = 3
		Expect(domain.ShowsCocoon(past)).To(BeFalse())
		Expect(domain.ShowsCocoon(monster.Monster{Genus: 35, Rank: 0})).To(BeFalse())
	})

	DescribeTable("is ready at 192 weeks, 80 loyalty, no fatigue and under 30 stress",
		func(change func(m *monster.Monster), want bool) {
			m := ready
			change(&m)
			Expect(domain.CocoonReady(m)).To(Equal(want))
		},
		Entry("all of it", func(*monster.Monster) {}, true),
		Entry("a week short", func(m *monster.Monster) { m.Age = 191 }, false),
		Entry("loyalty 79, with one of spoil and fear 79 and the other 80", func(m *monster.Monster) { m.Spoil, m.Fear = 79, 80 }, false),
		Entry("tired", func(m *monster.Monster) { m.Fatigue = 1 }, false),
		Entry("stressed", func(m *monster.Monster) { m.Stress = 30 }, false),
	)
})
