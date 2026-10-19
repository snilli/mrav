package demoui

import (
	"context"

	"fyne.io/fyne/v2/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
	"app/internal/infrastructure/fakemonster"
)

var _ = Describe("Panel", func() {
	var (
		src *fakemonster.Source
		p   *Panel
	)
	current := func() monster.Monster {
		m, err := src.Current(context.Background())
		Expect(err).NotTo(HaveOccurred())
		return m
	}

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		src = fakemonster.New()
		p = New(src, []domain.Genus{{Code: 0, Name: "Pixie"}, {Code: 7, Name: "Golem"}})
	})

	It("lists the genera by name and starts on the first", func() {
		Expect(p.genus.Options).To(Equal([]string{"Pixie", "Golem"}))
		Expect(p.genus.Selected).To(Equal("Pixie"))
	})

	It("sets the genus of the monster from the selection", func() {
		p.genus.SetSelected("Golem")
		Expect(current().Genus).To(Equal(monster.Genus(7)))
	})

	It("sets a stat as it is typed", func() {
		test.Type(p.stats[monster.Pow], "300")
		Expect(current().Stats[monster.Pow]).To(Equal(300))
		Expect(current().Stats[monster.Life]).To(Equal(0))
	})

	It("sets the nature, which can be negative, and the sub-genus", func() {
		test.Type(p.nature, "-20")
		test.Type(p.subGenus, "13")
		Expect(current().Nature).To(Equal(-20))
		Expect(current().SubGenus).To(Equal(monster.Genus(13)))
	})

	It("ignores text that is not a number instead of changing the monster", func() {
		test.Type(p.stats[monster.Pow], "30")
		p.stats[monster.Pow].SetText("3x")
		Expect(current().Stats[monster.Pow]).To(Equal(30))
	})

	It("teaches every move with the use count typed, and makes the first move of each column active", func() {
		test.Type(p.uses, "40")
		test.Tap(p.learnAll)
		m := current()
		for slot := 0; slot < monster.SlotCount; slot++ {
			Expect(m.Learned[slot]).To(BeTrue(), "slot %d", slot)
			Expect(m.Used[slot]).To(Equal(40), "slot %d", slot)
		}
		Expect(m.Active).To(Equal([monster.ActiveGroups]int{0, 6, 12, 18}))
	})

	It("makes the monster forget everything", func() {
		test.Tap(p.learnAll)
		test.Tap(p.forgetAll)
		m := current()
		Expect(m.Learned).To(Equal([monster.SlotCount]bool{}))
		Expect(m.Used).To(Equal([monster.SlotCount]int{}))
		Expect(m.Active).To(Equal([monster.ActiveGroups]int{monster.NoActiveMove, monster.NoActiveMove, monster.NoActiveMove, monster.NoActiveMove}))
	})
})
