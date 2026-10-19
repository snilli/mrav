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

	It("says how tired the monster is in the game's words, beside the number", func() {
		Expect(view(nil).FatigueWord).To(Equal("seems tired")) // a fatigue of 40
		Expect(view(func(p *domain.Profile) { p.Monster.Fatigue = 0 }).FatigueWord).To(Equal("very well"))
		Expect(view(func(p *domain.Profile) { p.Monster.Fatigue = 100 }).FatigueWord).To(Equal("pass out"))
	})

	It("sums up the monster's fights and its ailments", func() {
		v := view(func(p *domain.Profile) {
			p.Monster.Record = monster.Record{Wins: 12, Losses: 3, KOs: 5, Injuries: 4, Runaways: 6, Sicknesses: 7, Resistance: 8}
		})
		Expect(v.Record).To(Equal("12 W · 3 L · 5 KO"))
		Expect(v.History).To(Equal("sick 7 · injured 4 · ran away 6 · resistance 8"))
	})

	Describe("the ranch", func() {
		known := func(r player.Ranch, money int) func(*domain.Profile) {
			return func(p *domain.Profile) { p.Player.Ranch, p.Player.Money = r, money }
		}
		ranch := player.Ranch{Known: true, Date: 191, Grade: 6, House: 2, Barn: 1, Altar: 3}

		It("is not there when the game does not keep it", func() {
			Expect(view(nil).Ranch).To(Equal(profileui.RanchView{}))
		})

		It("says the date, with the game's own count of weeks to check it by, and how long to the upgrade day", func() {
			v := view(known(ranch, 1500)).Ranch
			Expect(v.Known).To(BeTrue())
			Expect(v.Date).To(Equal("Y1004 Mar wk 4 (week 191)"))
			Expect(v.UpgradeDay).To(Equal("May wk 4, in 8 weeks"))
		})

		It("says this week on the upgrade day, and the week before it in the singular", func() {
			day := ranch
			day.Date = 7
			Expect(view(known(day, 0)).Ranch.UpgradeDay).To(Equal("May wk 4, this week"))
			day.Date = 6
			Expect(view(known(day, 0)).Ranch.UpgradeDay).To(Equal("May wk 4, in 1 week"))
		})

		It("counts down to the IMA-FIMBA tournament where the year of the first was read, and is silent where it was not", func() {
			read := ranch
			read.Date, read.Tournament = 860, 1003
			Expect(view(known(read, 0)).Ranch.Tournament).To(Equal("Y1019 Jul wk 1, in 64 weeks"))
			read.Date = 924
			Expect(view(known(read, 0)).Ranch.Tournament).To(Equal("Y1019 Jul wk 1, this week"))
			Expect(view(known(ranch, 0)).Ranch.Tournament).To(BeEmpty())
		})

		It("lists the house, the stable and the altar: how far each is, and what its next step asks for", func() {
			v := view(known(ranch, 1500)).Ranch
			Expect(v.Facilities).To(Equal([]profileui.FacilityView{
				{
					Name: "House", Level: 2, Max: 4, Ratio: 0.5, Detail: "16 places",
					Next: "next 12,000 G · grade 6", State: profileui.Box{Text: "15,500 G more", Tone: profileui.ToneOrange},
				},
				{
					Name: "Stable", Level: 1, Max: 2, Ratio: 0.5,
					Next: "next 30,000 G · grade 7", State: profileui.Box{Text: "grade 7 · 33,500 G more", Tone: profileui.ToneOrange},
				},
				{
					Name: "Altar", Level: 3, Max: 3, Ratio: 1,
					Next: "fully upgraded", State: profileui.Box{Text: "MAX", Tone: profileui.ToneGray},
				},
			}))
		})

		It("says a step is ready when the grade is enough and the gold is the cost and 5,000 more", func() {
			short := view(known(ranch, 16999)).Ranch.Facilities[0].State
			Expect(short).To(Equal(profileui.Box{Text: "1 G more", Tone: profileui.ToneOrange}))
			ready := view(known(ranch, 17000)).Ranch.Facilities[0].State
			Expect(ready).To(Equal(profileui.Box{Text: "READY", Tone: profileui.ToneLimeGreen}))
		})

		It("lists the main breeds with whether each is unlocked, and counts the ones that are", func() {
			v := view(func(p *domain.Profile) {
				known(ranch, 1500)(p)
				p.Breeds = []domain.BreedUnlock{{Name: "Pixie", Unlocked: true}, {Name: "Dragon"}, {Name: "Worm", Unlocked: true}}
			}).Ranch
			Expect(v.Breeds).To(Equal([]profileui.BreedView{
				{Name: "Pixie", Unlocked: true}, {Name: "Dragon"}, {Name: "Worm", Unlocked: true},
			}))
			Expect(v.BreedsTitle).To(Equal("Breeds unlocked 2/3"))
		})

		It("has no list of breeds, and no title for it, when the game does not say which are unlocked", func() {
			v := view(known(ranch, 1500)).Ranch
			Expect(v.Breeds).To(BeEmpty())
			Expect(v.BreedsTitle).To(BeEmpty())
		})

		It("does not call the altar ready on gold alone, because it also waits 140 weeks that are not known", func() {
			basic := ranch
			basic.Altar = 1
			state := view(known(basic, 8000)).Ranch.Facilities[2].State
			Expect(state).To(Equal(profileui.Box{Text: "gold is enough", Tone: profileui.ToneLightGreen}))
		})

		It("asks for a grade the breeder lacks even with the gold in hand", func() {
			low := ranch
			low.Grade = 5
			state := view(known(low, 50000)).Ranch.Facilities[0].State
			Expect(state).To(Equal(profileui.Box{Text: "grade 6", Tone: profileui.ToneOrange}))
		})

		It("leaves the grade out of the altar's step, which asks for none, gives its own time and the gold the guide gives, and says when there is no altar yet", func() {
			basic := ranch
			basic.Altar = 1
			altar := view(known(basic, 1500)).Ranch.Facilities[2]
			Expect(altar.Next).To(Equal("next 5,000 G · Mar wk 3, 140 weeks after the last step"))
			Expect(altar.State).To(Equal(profileui.Box{Text: "6,500 G more", Tone: profileui.ToneOrange}))

			none := ranch
			none.Altar = 0
			Expect(view(known(none, 1500)).Ranch.Facilities[2]).To(Equal(profileui.FacilityView{
				Name: "Altar", Max: 3, Detail: "none", Next: "no altar yet",
			}))
		})
	})

	Describe("the expeditions", func() {
		dated := func(weeks int, energy int, energyKnown bool) func(*domain.Profile) {
			return func(p *domain.Profile) {
				p.Player.Ranch = player.Ranch{Known: true, Date: weeks, Energy: energy, EnergyKnown: energyKnown}
			}
		}
		names := func(v profileui.ExpeditionsView) []string {
			var out []string
			for _, t := range v.Trips {
				out = append(out, t.Name)
			}
			return out
		}

		It("lists the four trips in the order of the calendar when the date is not known, each with when it comes", func() {
			v := view(nil).Expeditions
			Expect(names(v)).To(Equal([]string{
				"Torles Mountains (Rovest)", "Phoenix Volcano Adventure", "Kawrea Volcano (Dr. Talico)", "Parepare Jungle (Kavaro)",
			}))
			Expect(v.Trips[0].When).To(Equal("Feb wk 1"))
			Expect(v.Trips[3].When).To(Equal("Oct wk 1"))
			Expect(v.Title).To(Equal("Next expedition: Torles Mountains (Rovest), Feb wk 1 · not fit yet"))
		})

		It("says what each invitation asks of the monster, and where it stands", func() {
			trip := view(nil).Expeditions.Trips[0] // the Pixie has life 111, fame 50 and grade E
			Expect(trip.Conditions).To(Equal([]profileui.ConditionView{
				{Text: "LIF 210 or more", Have: "111"},
				{Text: "Fame 50 or more", Have: "50", Met: true},
				{Text: "Grade C or higher", Have: "E"},
			}))
			Expect(trip.Fit).To(BeFalse())
		})

		It("calls a trip fit when the monster meets every condition", func() {
			strong := view(func(p *domain.Profile) {
				p.Monster.Stats[monster.Life], p.Monster.Fame, p.Monster.Rank = 300, 60, 4
			}).Expeditions
			for _, t := range strong.Trips {
				Expect(t.Fit).To(BeTrue(), t.Name)
			}
			Expect(strong.Title).To(HaveSuffix("· fit"))
		})

		It("puts the nearest first and says in how many weeks, when the date is known", func() {
			v := view(dated(191, 0, false)).Expeditions // March, week 4
			Expect(names(v)).To(Equal([]string{
				"Phoenix Volcano Adventure", "Kawrea Volcano (Dr. Talico)", "Parepare Jungle (Kavaro)", "Torles Mountains (Rovest)",
			}))
			Expect(v.Trips[0].When).To(Equal("Jun wk 1, in 9 weeks"))
			Expect(v.Trips[2].When).To(Equal("Oct wk 1, in 25 weeks"))
			Expect(v.Trips[3].When).To(Equal("Feb wk 1, in 41 weeks"))
			Expect(v.Title).To(Equal("Next expedition: Phoenix Volcano Adventure, Jun wk 1, in 9 weeks · not fit yet"))
		})

		It("says this week on the day, and the first expedition is the one that is made once", func() {
			v := view(dated(8, 0, false)).Expeditions // June, week 1
			Expect(v.Trips[0].When).To(Equal("Jun wk 1, this week"))
			Expect(v.Trips[0].Note).To(Equal("the first expedition: made once, and it unlocks the others"))
			Expect(v.Trips[1].Note).To(BeEmpty())
		})

		It("says the energy the monster would start with, from its life", func() {
			v := view(func(p *domain.Profile) { p.Monster.Stats[monster.Life] = 305 }).Expeditions
			Expect(v.StartEnergy).To(Equal("Starts with 280 energy (70 for every 70 LIF, from 140 up to 700)"))
		})

		It("says the energy in the game now only when it was read", func() {
			Expect(view(dated(0, -5, true)).Expeditions.GameEnergy).To(Equal("Energy in the game now: -5 (it stays after an expedition)"))
			Expect(view(dated(0, 0, false)).Expeditions.GameEnergy).To(BeEmpty())
			Expect(view(nil).Expeditions.GameEnergy).To(BeEmpty())
		})

		It("gives what an expedition costs, and what failing costs", func() {
			Expect(view(nil).Expeditions.Costs).To(Equal("A step costs 1 energy, an obstacle 5, a search 10; ending under -9 costs 12 weeks of life"))
		})
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
