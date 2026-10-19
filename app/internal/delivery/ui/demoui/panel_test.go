package demoui

import (
	"context"
	"math/rand/v2"
	"strconv"

	"fyne.io/fyne/v2/test"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
	"app/internal/infrastructure/fakemonster"
	"app/internal/infrastructure/fakeplayer"
)

var _ = Describe("Panel", func() {
	var (
		src     *fakemonster.Source
		players *fakeplayer.Source
		p       *Panel
	)
	current := func() monster.Monster {
		m, err := src.Current(context.Background())
		Expect(err).NotTo(HaveOccurred())
		return m
	}

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		src = fakemonster.New()
		players = fakeplayer.New()
		p = New(src, players, []domain.Genus{{Code: 0, Name: "Pixie"}, {Code: 7, Name: "Golem"}})
		p.rng = rand.New(rand.NewPCG(1, 2))
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
		for slot := range monster.SlotCount {
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

	It("sets the other numbers of the monster as they are typed, a negative form included", func() {
		test.Type(p.age, "100")
		test.Type(p.lifespan, "250")
		test.Type(p.initial, "300")
		test.Type(p.fatigue, "40")
		test.Type(p.stress, "60")
		test.Type(p.fame, "50")
		test.Type(p.form, "-10")
		test.Type(p.spoil, "70")
		test.Type(p.fear, "80")
		test.Type(p.lifeStage, "5")
		test.Type(p.growth[monster.Pow], "3")
		m := current()
		Expect([]int{m.Age, m.Lifespan, m.InitialLifespan, m.Fatigue, m.Stress, m.Fame, m.Form, m.Spoil, m.Fear, m.LifeStage}).
			To(Equal([]int{100, 250, 300, 40, 60, 50, -10, 70, 80, 5}))
		Expect(m.Growth[monster.Pow]).To(Equal(3))
		Expect(m.Growth[monster.Life]).To(Equal(0))
	})

	It("sets the money of the player as it is typed", func() {
		test.Type(p.money, "1500")
		pl, err := players.Player(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect(pl.Money).To(Equal(1500))
	})

	It("randomizes the monster and the player, and shows the new numbers in the boxes", func() {
		test.Tap(p.random)

		m := current()
		pl, err := players.Player(context.Background())
		Expect(err).NotTo(HaveOccurred())
		Expect([]monster.Genus{0, 7}).To(ContainElement(m.Genus))
		Expect(m.Stats[monster.Pow]).To(BeNumerically(">=", 1))
		Expect(pl.Breeder).To(Equal("Demo Breeder"))

		Expect([]string{"Pixie", "Golem"}).To(ContainElement(p.genus.Selected))
		Expect(p.stats[monster.Pow].Text).To(Equal(strconv.Itoa(m.Stats[monster.Pow])))
		Expect(p.age.Text).To(Equal(strconv.Itoa(m.Age)))
		Expect(p.growth[monster.Life].Text).To(Equal(strconv.Itoa(m.Growth[monster.Life])))
		Expect(p.form.Text).To(Equal(strconv.Itoa(m.Form)))
		Expect(p.money.Text).To(Equal(strconv.Itoa(pl.Money)))
	})
})
