package demoui

import (
	"math/rand/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
	player "app/internal/core/player/domain"
)

var _ = Describe("randomize", func() {
	genera := []domain.Genus{{Code: 0, Name: "Pixie"}, {Code: 7, Name: "Golem"}, {Code: 37, Name: "Naga"}}
	roll := func(seed uint64) (monster.Monster, player.Player) {
		var m monster.Monster
		var p player.Player
		randomize(rand.New(rand.NewPCG(seed, 0)), genera, &m, &p)
		return m, p
	}

	It("gives the same monster and player for the same seed, and others for others", func() {
		m1, p1 := roll(1)
		m2, p2 := roll(1)
		Expect(m1).To(Equal(m2))
		Expect(p1).To(Equal(p2))

		m3, _ := roll(2)
		Expect(m3).NotTo(Equal(m1))
	})

	It("gives only what the game could have, whatever the seed", func() {
		for seed := range uint64(300) {
			m, p := roll(seed)
			note := func() string { return "seed " + string(rune('0'+seed%10)) }
			Expect([]monster.Genus{0, 7, 37}).To(ContainElement(m.Genus), note())
			Expect(m.SubGenus).To(BeNumerically("<", 38))
			for stat, v := range m.Stats {
				Expect(v).To(BeNumerically(">=", 1), "stat %d", stat)
				Expect(v).To(BeNumerically("<=", 999), "stat %d", stat)
				Expect(m.Growth[stat]).To(BeNumerically("<=", 4))
			}
			Expect(m.GutsRate).To(BeNumerically(">=", 5)) // nothing less, or the game shows no monster
			Expect(m.Lifespan).To(BeNumerically(">=", 1))
			Expect(m.Lifespan).To(BeNumerically("<=", m.InitialLifespan))
			Expect(m.Specials).To(BeNumerically("<", 1<<10)) // the ordinary specials, none that a breed cannot have
			Expect(m.ItemLike).To(BeNumerically("<", 178))
			Expect(m.ItemDislike).To(BeNumerically("<", 178))
			for _, percent := range m.Motivation {
				Expect(percent).To(BeNumerically("<=", 100))
			}
			for c, slot := range m.Active { // each column's active move is in the column, or there is none
				Expect(slot == monster.NoActiveMove || slot >= c*domain.Rows && slot < (c+1)*domain.Rows).To(BeTrue(), "column %d", c)
			}
			for _, id := range p.Items {
				Expect(id).To(BeNumerically("<=", 178))
			}
			Expect(p.Breeder).NotTo(BeEmpty())
			Expect(m.Name).NotTo(BeEmpty())
			// a ranch that the game could have: levels within what each facility goes to, a grade from Beginner to Master
			Expect(p.Ranch.Known).To(BeTrue())
			Expect(p.Ranch.House).To(BeNumerically("<=", player.MaxLevel(player.House)))
			Expect(p.Ranch.Barn).To(BeNumerically("<=", player.MaxLevel(player.Barn)))
			Expect(p.Ranch.Altar).To(BeNumerically("<=", player.MaxLevel(player.Altar)))
			Expect(p.Ranch.Grade).To(BeNumerically("<=", 10))
			Expect(p.Ranch.Date).To(BeNumerically(">=", 0))
			Expect(m.Record.Wins).To(BeNumerically(">=", 0))
		}
	})

	It("makes most monsters pure breeds, which every breed table has, and some crossbreeds", func() {
		pure, total := 0, 400
		for seed := range uint64(total) {
			if m, _ := roll(seed); m.SubGenus == m.Genus {
				pure++
			}
		}
		Expect(pure).To(BeNumerically(">", total*6/10))
		Expect(pure).To(BeNumerically("<", total))
	})

	It("leaves the genus alone when there are no genera to choose from", func() {
		m := monster.Monster{Genus: 9}
		var p player.Player
		randomize(rand.New(rand.NewPCG(1, 0)), nil, &m, &p)
		Expect(m.Genus).To(Equal(monster.Genus(9)))
	})
})
