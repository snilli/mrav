package fakemonster_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/infrastructure/fakemonster"
	"app/internal/port"
)

var _ = Describe("Source", func() {
	ctx := context.Background()

	It("is a MonsterSource", func() {
		var _ port.MonsterSource = fakemonster.New()
	})

	It("starts with a Pixie that has no active move", func() {
		m, err := fakemonster.New().Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(m.Genus).To(Equal(monster.Genus(0)))
		Expect(m.Active).To(Equal([monster.ActiveGroups]int{monster.NoActiveMove, monster.NoActiveMove, monster.NoActiveMove, monster.NoActiveMove}))
	})

	It("comes from the Steam game, the version that the shipped moves are based on", func() {
		m, err := fakemonster.New().Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(m.Version).To(Equal(monster.MR2DX))
	})

	It("shows what Update set", func() {
		s := fakemonster.New()
		s.Update(func(m *monster.Monster) {
			m.Genus = 7
			m.Stats[monster.Pow] = 300
			m.Learned[2] = true
		})
		m, err := s.Current(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(m.Genus).To(Equal(monster.Genus(7)))
		Expect(m.Stats[monster.Pow]).To(Equal(300))
		Expect(m.Learned[2]).To(BeTrue())
	})

	It("hands out a copy, so a reader cannot change the monster", func() {
		s := fakemonster.New()
		m, _ := s.Current(ctx)
		m.Stats[monster.Life] = 999
		again, _ := s.Current(ctx)
		Expect(again.Stats[monster.Life]).To(Equal(0))
	})
})
