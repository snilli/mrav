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
	player "app/internal/core/player/domain"
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

	It("sets the rank of the monster from the letter selected, E to S", func() {
		Expect(p.rank.Options).To(Equal([]string{"E", "D", "C", "B", "A", "S"}))
		p.rank.SetSelected("B")
		Expect(current().Rank).To(Equal(3))
		p.rank.SetSelected("S")
		Expect(current().Rank).To(Equal(5))
	})

	It("moves the date of the ranch to the month and the week selected, in the year that it is in, and makes the ranch known", func() {
		players.Update(func(pl *player.Player) { pl.Ranch = player.Ranch{Known: true, Date: 191} })
		p.month.SetSelected("Feb")
		pl, _ := players.Player(context.Background())
		Expect(pl.Ranch.Calendar()).To(Equal(player.Date{Year: 1004, Month: 2, Week: 4}), "the week is left as it was")
		p.week.SetSelected("1")
		pl, _ = players.Player(context.Background())
		Expect(pl.Ranch.Calendar()).To(Equal(player.Date{Year: 1004, Month: 2, Week: 1}), "and the month")

		players.Update(func(pl *player.Player) { pl.Ranch = player.Ranch{} })
		p.week.SetSelected("2")
		pl, _ = players.Player(context.Background())
		Expect(pl.Ranch.Known).To(BeTrue())
	})

	It("puts the monster on the map of a trip as far right and up of the start as typed, and sets its energy left", func() {
		test.Type(p.tripX, "-3")
		test.Type(p.tripY, "12")
		test.Type(p.energy, "-5")
		pl, _ := players.Player(context.Background())
		x, y := pl.Ranch.Walked()
		Expect([]int{x, y}).To(Equal([]int{-3, 12}))
		Expect(pl.Ranch.WalkKnown).To(BeTrue())
		Expect(pl.Ranch.Energy).To(Equal(-5), "which the game can take below nothing")
		Expect(pl.Ranch.EnergyKnown).To(BeTrue())
	})

	It("shows where a random monster is on the map, with its energy, in the boxes", func() {
		p.Randomize()
		pl, _ := players.Player(context.Background())
		Expect(p.tripX.Text).To(Equal("0"))
		Expect(p.tripY.Text).To(Equal("0"))
		Expect(p.energy.Text).To(Equal(strconv.Itoa(pl.Ranch.Energy)))
	})

	It("shows the rank and the date of a random monster in the boxes", func() {
		p.Randomize()
		pl, _ := players.Player(context.Background())
		cal := pl.Ranch.Calendar()
		Expect(p.rank.Selected).To(Equal("EDCBAS"[current().Rank : current().Rank+1]))
		Expect(p.month.Selected).To(Equal(player.MonthNames[cal.Month-1]))
		Expect(p.week.Selected).To(Equal(strconv.Itoa(cal.Week)))
		Expect(pl.Ranch.Calendar()).To(Equal(cal), "showing them did not move the date")
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
