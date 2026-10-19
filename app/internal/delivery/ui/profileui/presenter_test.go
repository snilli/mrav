package profileui_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
)

var _ = Describe("NewMonsterView", func() {
	// a Pixie, mid-life, with every number different so that a mix-up shows
	pixie := func() domain.Profile {
		m := monster.Monster{
			Genus: 12, SubGenus: 3, Name: "Aa 1",
			Stats:  [monster.StatCount]int{111, 222, 666, 444, 555, 333}, // Life, Pow, Int, Skill, Speed, Defense
			Nature: 130, NatureBase: 30,
			Age: 100, Lifespan: 250, InitialLifespan: 300,
			Fatigue: 40, Stress: 60, Fame: 50, Form: -10, Spoil: 70, Fear: 80,
			Growth:   [monster.StatCount]int{3, 1, 2, 4, 0, 2},
			PlayType: 2, LifeStage: 5, LifeType: 3, ArenaSpeed: 4, GutsRate: 5, PrizeMoney: 12345,
		}
		return domain.Profile{
			Monster: m, Player: player.Player{Money: 1500, NextSale: 3, ErrantryCD: 255},
			Breed: domain.Breed{Name: "Pixie", GutsOK: true}, MainGenus: "Pixie", SubGenus: "Dragon",
			Like: "Cup Jelly", Dislike: "Nothing", Drug: "---",
		}
	}
	view := func(change func(*domain.Profile)) profileui.MonsterView {
		p := pixie()
		if change != nil {
			change(&p)
		}
		return profileui.NewMonsterView(p, profileui.Options{})
	}

	It("names the monster and its breeds, and gives its age both ways", func() {
		v := view(nil)
		Expect(v.Breed).To(Equal(profileui.Box{Text: "Pixie"}))
		Expect(v.GivenName).To(Equal("Aa 1"))
		Expect(v.MainGenus).To(Equal("Pixie"))
		Expect(v.SubGenus).To(Equal("Dragon"))
		Expect(v.Age).To(Equal("2y,1m,0w (100w)"))
		Expect(view(func(p *domain.Profile) { p.Monster.Age = -1 }).Age).To(Equal("---"))
	})

	It("flags a breed name that is not the standard one", func() {
		Expect(view(func(p *domain.Profile) { p.Breed.Name = "Pixie (N/S)" }).Breed.Tone).To(Equal(profileui.ToneLightPink))
		Expect(view(func(p *domain.Profile) { p.Breed.Name = "[E] Silver Face" }).Breed.Tone).To(Equal(profileui.ToneLightPink))
	})

	It("shows the stages and the lifespan in words", func() {
		v := view(nil)
		Expect(v.LifeStage).To(Equal("6 - Sub-Prime"))
		Expect(v.LifeType).To(Equal("Type 4: Sustainable"))
		Expect(v.Lifespan).To(Equal("250w"))
		Expect(v.InitialLifespan).To(Equal("300w"))
		Expect(view(func(p *domain.Profile) { p.Monster.Lifespan = 0 }).LifeStage).To(Equal("Dead :("))
	})

	It("shows the stats, with the effective speed and defense after the form beside them, and their total", func() {
		v := view(nil)
		Expect(v.Stats).To(Equal([monster.StatCount]string{"111", "222", "666", "444", "555 [568]", "333 [324]"}))
		Expect(v.StatTotal).To(Equal("2331"))
	})

	It("sizes the bars of the stats, the condition, the lifespan and the life index, and colours the ones that warn", func() {
		v := view(nil)
		for stat, value := range []int{111, 222, 666, 444, 555, 333} {
			Expect(v.Bars.Stats[stat]).To(BeNumerically("~", float64(value)/999, 1e-9), "stat %d", stat)
		}
		Expect(v.Bars.Fatigue.Ratio).To(BeNumerically("~", 0.4, 1e-9))
		Expect(v.Bars.Fatigue.Tone).To(Equal(profileui.ToneOrange))
		Expect(v.Bars.Stress.Ratio).To(BeNumerically("~", 0.6, 1e-9))
		Expect(v.Bars.Stress.Tone).To(Equal(profileui.ToneOrange))
		Expect(v.Bars.Lifespan.Ratio).To(BeNumerically("~", 250.0/300, 1e-9))
		Expect(v.Bars.Lifespan.Tone).To(Equal(profileui.ToneLimeGreen))
		Expect(v.Bars.LifeIndex.Ratio).To(BeNumerically("~", 160.0/280, 1e-9))
		Expect(v.Bars.LifeIndex.Tone).To(Equal(profileui.ToneOrange)) // the colour of its box: 3 weeks lost
	})

	DescribeTable("colours fatigue and stress green under 40, orange under 70, and red from there",
		func(value int, want profileui.Tone) {
			v := view(func(p *domain.Profile) { p.Monster.Fatigue, p.Monster.Stress = value, value })
			Expect(v.Bars.Fatigue.Tone).To(Equal(want))
			Expect(v.Bars.Stress.Tone).To(Equal(want))
		},
		Entry("none", 0, profileui.ToneLimeGreen),
		Entry("39", 39, profileui.ToneLimeGreen),
		Entry("40", 40, profileui.ToneOrange),
		Entry("69", 69, profileui.ToneOrange),
		Entry("70", 70, profileui.ToneRed),
		Entry("100", 100, profileui.ToneRed),
	)

	DescribeTable("puts the nature on a gauge from -100 to 100, and keeps it on the gauge",
		func(nature int, want float64) {
			v := view(func(p *domain.Profile) { p.Monster.Nature = nature })
			Expect(v.Bars.Nature.Ratio).To(BeNumerically("~", want, 1e-9))
		},
		Entry("the lowest", -100, 0.0),
		Entry("the middle", 0, 0.5),
		Entry("the highest", 100, 1.0),
		Entry("past the top", 130, 1.0),
		Entry("past the bottom", -250, 0.0),
	)

	DescribeTable("colours the lifespan left: green over half, orange over a quarter, then red",
		func(left, initial int, want profileui.Tone, ratio float64) {
			v := view(func(p *domain.Profile) { p.Monster.Lifespan, p.Monster.InitialLifespan = left, initial })
			Expect(v.Bars.Lifespan.Tone).To(Equal(want))
			Expect(v.Bars.Lifespan.Ratio).To(BeNumerically("~", ratio, 1e-9))
		},
		Entry("all of it", 300, 300, profileui.ToneLimeGreen, 1.0),
		Entry("half is not over half", 150, 300, profileui.ToneOrange, 0.5),
		Entry("a quarter is not over a quarter", 75, 300, profileui.ToneRed, 0.25),
		Entry("none left", 0, 300, profileui.ToneRed, 0.0),
		Entry("no initial lifespan to measure by", 50, 0, profileui.ToneRed, 0.0),
		Entry("more than it began with", 320, 300, profileui.ToneLimeGreen, 1.0),
	)

	It("shows each growth rate as its grade and its number from 1, or from 0", func() {
		v := view(nil)
		Expect(v.Growth[monster.Life]).To(Equal(profileui.Box{Text: "B / 4", Tone: profileui.ToneYellowGreen}))
		Expect(v.Growth[monster.Pow]).To(Equal(profileui.Box{Text: "D / 2", Tone: profileui.ToneOrange}))
		Expect(v.Growth[monster.Speed]).To(Equal(profileui.Box{Text: "E / 1", Tone: profileui.ToneOrangeRed}))
		Expect(v.Growth[monster.Skill]).To(Equal(profileui.Box{Text: "A / 5", Tone: profileui.ToneLimeGreen}))
		Expect(view(func(p *domain.Profile) { p.Monster.Growth[monster.Int] = 5 }).Growth[monster.Int]).
			To(Equal(profileui.Box{Text: "S / 6", Tone: profileui.ToneHotPink}))
		Expect(view(func(p *domain.Profile) { p.Monster.Growth[monster.Int] = 11 }).Growth[monster.Int]).
			To(Equal(profileui.Box{Text: "--- / 12"}))

		p := pixie()
		zero := profileui.NewMonsterView(p, profileui.Options{GrowthFromZero: true})
		Expect(zero.Growth[monster.Life].Text).To(Equal("B / 3"))
	})

	It("shows loyalty with its two halves, fame, stress and fatigue, and the form in words", func() {
		v := view(nil)
		Expect(v.Loyalty).To(Equal("75"))
		Expect(v.Spoil).To(Equal("70"))
		Expect(v.Fear).To(Equal("80"))
		Expect(v.Fame).To(Equal(profileui.Box{Text: "50"}))
		Expect(v.Stress).To(Equal("60"))
		Expect(v.Fatigue).To(Equal("40"))
		Expect(v.Form).To(Equal("-10 (Normal)"))
		Expect(view(func(p *domain.Profile) { p.Monster.Form = -20 }).Form).To(Equal("-20 (Slim)"))
		Expect(view(func(p *domain.Profile) { p.Monster.Fame = 101 }).Fame.Tone).To(Equal(profileui.ToneHotPink))
	})

	DescribeTable("shows the life index with the weeks it costs, and a colour that warns",
		func(fatigue, stress int, text string, tone profileui.Tone) {
			v := view(func(p *domain.Profile) { p.Monster.Fatigue, p.Monster.Stress = fatigue, stress })
			Expect(v.LifeIndex).To(Equal(profileui.Box{Text: text, Tone: tone}))
		},
		Entry("nothing lost", 0, 0, "0 (-0 Weeks)", profileui.ToneLightGreen),
		Entry("one week", 70, 0, "70 (-1 Week)", profileui.ToneLightYellow),
		Entry("two weeks", 105, 0, "105 (-2 Weeks)", profileui.ToneLightYellow),
		Entry("three weeks", 140, 0, "140 (-3 Weeks)", profileui.ToneOrange),
		Entry("four weeks", 175, 0, "175 (-4 Weeks)", profileui.ToneOrange),
		Entry("five weeks", 210, 0, "210 (-5 Weeks)", profileui.ToneOrangeRed),
		Entry("six weeks", 245, 0, "245 (-6 Weeks)", profileui.ToneOrangeRed),
		Entry("seven weeks", 280, 0, "280 (-7 Weeks)", profileui.ToneRed),
	)

	It("shows the nature base, and the effective nature with its label", func() {
		v := view(nil)
		Expect(v.NatureBase).To(Equal("30"))
		Expect(v.Nature).To(Equal(profileui.Box{Text: "130 (Best)"}))
		Expect(view(func(p *domain.Profile) { p.Monster.Nature = 0 }).Nature).To(Equal(profileui.Box{Text: "0 (Neutral)"}))
		Expect(view(func(p *domain.Profile) { p.Monster.Nature = 30 }).Nature).To(Equal(profileui.Box{Text: "30 (Good)", Tone: profileui.ToneLightYellow}))
		Expect(view(func(p *domain.Profile) { p.Monster.Nature = -30 }).Nature).To(Equal(profileui.Box{Text: "-30 (Bad)", Tone: profileui.ToneGray}))
		Expect(view(func(p *domain.Profile) { p.Monster.Nature = -80 }).Nature).To(Equal(profileui.Box{Text: "-80 (Worst)", Tone: profileui.ToneBlack}))
	})

	It("shows the guts rate, flagged when it is not the breed's, and as guts a second when asked", func() {
		Expect(view(nil).GutsRate).To(Equal(profileui.Box{Text: "5"}))
		Expect(view(func(p *domain.Profile) { p.Breed.GutsOK = false }).GutsRate.Tone).To(Equal(profileui.ToneHotPink))
		per := profileui.NewMonsterView(pixie(), profileui.Options{GutsPerSecond: true})
		Expect(per.GutsRate.Text).To(Equal("6.00"))
	})

	It("grades the arena speed", func() {
		Expect(view(nil).ArenaSpeed).To(Equal(profileui.Box{Text: "A", Tone: profileui.ToneLimeGreen}))
		Expect(view(func(p *domain.Profile) { p.Monster.ArenaSpeed = 0 }).ArenaSpeed).To(Equal(profileui.Box{Text: "E", Tone: profileui.ToneOrangeRed}))
		Expect(view(func(p *domain.Profile) { p.Monster.ArenaSpeed = 9 }).ArenaSpeed).To(Equal(profileui.Box{Text: "---"}))
	})

	It("lists the battle specials, and flags a set that the breed cannot have", func() {
		Expect(view(nil).Specials).To(Equal(profileui.Box{}))
		v := view(func(p *domain.Profile) { p.Monster.Specials = 1 | 8 })
		Expect(v.Specials).To(Equal(profileui.Box{Text: "[Power], Will"}))
		v = view(func(p *domain.Profile) { p.Monster.Specials = 1024 })
		Expect(v.Specials).To(Equal(profileui.Box{Text: "Real", Tone: profileui.ToneHotPink}))
	})

	It("shows what the monster wants and does not want, in colour when it is a Tablet or a Battle", func() {
		v := view(nil)
		Expect(v.Like).To(Equal(profileui.Box{Text: "Cup Jelly"}))
		Expect(v.Dislike).To(Equal(profileui.Box{Text: "Nothing"}))
		v = view(func(p *domain.Profile) { p.Like, p.Dislike = "Tablet", "Battle" })
		Expect(v.Like).To(Equal(profileui.Box{Text: "Tablet", Tone: profileui.ToneLightGreen}))
		Expect(v.Dislike).To(Equal(profileui.Box{Text: "Battle", Tone: profileui.ToneOrangeRed}))
	})

	It("names the play", func() {
		Expect(view(nil).PlayType).To(Equal("Sparring"))
	})

	It("names the drug with its weeks and its tip, or shows none", func() {
		v := view(nil)
		Expect(v.Drug).To(Equal("---"))
		Expect(v.DrugWeeks).To(BeEmpty())
		Expect(v.DrugTip).To(BeEmpty())

		v = view(func(p *domain.Profile) { p.Monster.Drug, p.Monster.DrugWeeks, p.Drug = 38, 3, "Troron" })
		Expect(v.Drug).To(Equal("Troron"))
		Expect(v.DrugWeeks).To(Equal("(3w)"))
		Expect(v.DrugTip).To(HavePrefix("Troron:"))
	})

	It("shows the peaches, the item used, the money, the prize money, the next sale and the errantry sale", func() {
		v := view(func(p *domain.Profile) { p.Monster.GoldPeach, p.Monster.ItemUsed = true, true })
		Expect(v.GoldPeach).To(BeTrue())
		Expect(v.SilverPeach).To(BeFalse())
		Expect(v.ItemUsed).To(BeTrue())
		Expect(v.Money).To(Equal("1500G"))
		Expect(v.PrizeMoney).To(Equal("12345G"))
		Expect(v.NextSale).To(Equal("3w"))
		Expect(v.ErrantrySale).To(BeTrue())
		Expect(view(func(p *domain.Profile) { p.Player.ErrantryCD = 7 }).ErrantrySale).To(BeFalse())
	})

	It("shows the cocoon facts for a young Worm only", func() {
		Expect(view(nil).Cocoon).To(BeNil())

		worm := func(p *domain.Profile) {
			p.Monster.Genus, p.Monster.Rank, p.Monster.CupJellies = 36, 2, 30
			p.Monster.Age, p.Monster.Spoil, p.Monster.Fear, p.Monster.Fatigue, p.Monster.Stress = 192, 80, 80, 0, 29
		}
		Expect(view(worm).Cocoon).To(Equal(&profileui.CocoonView{CupJellies: profileui.Box{Text: "30", Tone: profileui.ToneLightGreen}, Ready: true}))
		few := view(func(p *domain.Profile) { worm(p); p.Monster.CupJellies = 29; p.Monster.Age = 100 })
		Expect(few.Cocoon).To(Equal(&profileui.CocoonView{CupJellies: profileui.Box{Text: "29"}, Ready: false}))
		Expect(view(func(p *domain.Profile) { worm(p); p.Monster.Rank = 3 }).Cocoon).To(BeNil())
	})
})

var _ = DescribeTable("MotivationTone says how eager a monster is to train",
	func(percent int, tone profileui.Tone) { Expect(profileui.MotivationTone(percent)).To(Equal(tone)) },
	Entry("none at all", 0, profileui.ToneOrangeRed),
	Entry("just under the middle", 29, profileui.ToneOrangeRed),
	Entry("the middle begins", 30, profileui.ToneLightYellow),
	Entry("the middle ends", 80, profileui.ToneLightYellow),
	Entry("the high begins", 81, profileui.ToneLightGreen),
	Entry("full", 100, profileui.ToneLightGreen),
	Entry("past full, which the game does not give", 101, profileui.ToneNone),
)
