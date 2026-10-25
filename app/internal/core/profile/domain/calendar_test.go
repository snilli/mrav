package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
)

var _ = Describe("Calendar", func() {
	const may4of1001, jun1of1001 = 55, 56
	rank := func(letter string) int {
		for r := range 6 {
			if domain.RankLetter(r) == letter {
				return r
			}
		}
		panic(letter)
	}
	when := func(t domain.CondType, v int) domain.EntryCondition { return domain.EntryCondition{Type: t, Value: v} }
	profile := func(date int, m monster.Monster, entries []domain.CalendarEntry, change ...func(*player.Player)) domain.Profile {
		p := domain.Profile{Monster: m, Calendar: entries, Player: player.Player{
			NextSale: 99, Ranch: player.Ranch{Known: true, Date: date, Grade: 1, PhoenixKnown: true, PhoenixDone: true},
		}}
		for _, c := range change {
			c(&p.Player)
		}
		return p
	}
	fit := func(life, fame int, letter string) monster.Monster {
		m := monster.Monster{Fame: fame, Rank: rank(letter)}
		m.Stats[monster.Life] = life
		return m
	}
	on := func(weeks []domain.CalendarWeek, year, month, week int) domain.CalendarWeek {
		for _, w := range weeks {
			if w.Date == (player.Date{Year: year, Month: month, Week: week}) {
				return w
			}
		}
		Fail("no such week")
		return domain.CalendarWeek{}
	}
	named := func(w domain.CalendarWeek, name string) (domain.Event, bool) {
		for _, e := range w.Events {
			if e.Name == name {
				return e, true
			}
		}
		return domain.Event{}, false
	}
	event := func(w domain.CalendarWeek, name string) domain.Event {
		e, ok := named(w, name)
		Expect(ok).To(BeTrue(), name)
		return e
	}
	one := func(m monster.Monster, e domain.CalendarEntry, change ...func(*player.Player)) domain.Event {
		e.Month, e.Week, e.Name = 7, 2, "Cup"
		return event(on(domain.Calendar(profile(jun1of1001, m, []domain.CalendarEntry{e}, change...)), 1001, 7, 2), "Cup")
	}

	It("runs week by week from this week to the end of next year, counting the weeks from now", func() {
		weeks := domain.Calendar(profile(jun1of1001, fit(0, 0, "E"), nil))
		Expect(weeks[0].Date).To(Equal(player.Date{Year: 1001, Month: 6, Week: 1}))
		Expect(weeks[len(weeks)-1].Date).To(Equal(player.Date{Year: 1002, Month: 12, Week: 4}))
		for i, w := range weeks {
			Expect(w.In).To(Equal(i))
		}
		Expect(weeks).To(HaveLen(7*4 + 12*4))
	})

	It("has nothing when the game does not say the date", func() {
		Expect(domain.Calendar(domain.Profile{})).To(BeEmpty())
	})

	It("puts an entry in its week of every year, with its kind, and its purse and prize as the detail", func() {
		entry := domain.CalendarEntry{Month: 2, Week: 2, Name: "Poannka Cup (B)", Kind: domain.EventCup, Purse: 5000, Prize: "Disc Chips (Hopper)"}
		weeks := domain.Calendar(profile(0, fit(0, 0, "B"), []domain.CalendarEntry{entry}))
		want := domain.Event{Name: "Poannka Cup (B)", Kind: domain.EventCup, Fit: domain.FitReady, Detail: "5000G, Disc Chips (Hopper)"}
		Expect(event(on(weeks, 1001, 2, 2), "Poannka Cup (B)")).To(Equal(want))
		_, ok := named(on(weeks, 1001, 2, 3), "Poannka Cup (B)")
		Expect(ok).To(BeFalse())
	})

	It("lists the tournaments of a week first, from grade S down to F, and the other events after them", func() {
		at := func(name string, kind domain.EventKind, when ...domain.EntryCondition) domain.CalendarEntry {
			return domain.CalendarEntry{Month: 2, Week: 2, Name: name, Kind: kind, When: when}
		}
		entries := []domain.CalendarEntry{
			at("A trip", domain.EventTrip),
			at("Torles Tourney (D)", domain.EventCup, when(domain.CondCupGrade, 1)),
			at("New Year Cup (F)", domain.EventCup),
			at("Poannka Cup (B)", domain.EventCup, when(domain.CondCupGrade, 3)),
			at("Greatest 4 (S)", domain.EventCup, when(domain.CondRankAtLeast, 5)),
		}
		var names []string
		for _, e := range on(domain.Calendar(profile(0, fit(0, 0, "E"), entries)), 1001, 2, 2).Events {
			names = append(names, e.Name)
		}
		Expect(names).To(Equal([]string{"Greatest 4 (S)", "Poannka Cup (B)", "Torles Tourney (D)", "New Year Cup (F)", "A trip"}))
	})

	DescribeTable("weighs each condition: met, not yet with what it needs, or never",
		func(m monster.Monster, c domain.EntryCondition, want domain.Fit, need string) {
			e := one(m, domain.CalendarEntry{When: []domain.EntryCondition{c}})
			Expect(e.Fit).To(Equal(want))
			Expect(e.Need).To(Equal(need))
		},
		Entry("rank at least, met", fit(0, 0, "S"), when(domain.CondRankAtLeast, 3), domain.FitReady, ""),
		Entry("rank at least, below", fit(0, 0, "C"), when(domain.CondRankAtLeast, 3), domain.FitNotYet, "Rank B"),
		Entry("rank exactly, met", fit(0, 0, "B"), when(domain.CondRankExactly, 3), domain.FitReady, ""),
		Entry("rank exactly, below", fit(0, 0, "C"), when(domain.CondRankExactly, 3), domain.FitNotYet, "Rank B"),
		Entry("rank exactly, past it for good", fit(0, 0, "A"), when(domain.CondRankExactly, 3), domain.FitNever, ""),
		Entry("cup grade, met", fit(0, 0, "B"), when(domain.CondCupGrade, 3), domain.FitReady, ""),
		Entry("cup grade, below", fit(0, 0, "C"), when(domain.CondCupGrade, 3), domain.FitNotYet, "Rank B"),
		Entry("cup grade, past it: shown with a warning", fit(0, 0, "A"), when(domain.CondCupGrade, 3), domain.FitWarn, "Rank above B"),
		Entry("age under, met", monster.Monster{Age: 23}, when(domain.CondAgeUnder, 24), domain.FitReady, ""),
		Entry("age under, too old for good", monster.Monster{Age: 24}, when(domain.CondAgeUnder, 24), domain.FitNever, ""),
		Entry("age at least, too young", monster.Monster{Age: 299}, when(domain.CondAgeAtLeast, 300), domain.FitNotYet, "Age 300 weeks"),
		Entry("nature at least, met", monster.Monster{Nature: 20}, when(domain.CondNatureAtLeast, 20), domain.FitReady, ""),
		Entry("nature at least, short", monster.Monster{Nature: 0}, when(domain.CondNatureAtLeast, 20), domain.FitNotYet, "Nature 20 or more"),
		Entry("nature at most, short", monster.Monster{Nature: 0}, when(domain.CondNatureAtMost, -20), domain.FitNotYet, "Nature -20 or less"),
	)

	It("asks for the trainer's rank and the year", func() {
		e := one(fit(0, 0, "E"), domain.CalendarEntry{When: []domain.EntryCondition{when(domain.CondTrainerRankAtLeast, 6), when(domain.CondYearAtLeast, 1010)}})
		Expect(e.Fit).To(Equal(domain.FitNotYet))
		Expect(e.Need).To(Equal("Trainer rank 6, Year 1010"))
	})

	It("is never when one condition is never, whatever the others say", func() {
		e := one(fit(0, 0, "A"), domain.CalendarEntry{When: []domain.EntryCondition{when(domain.CondYearAtLeast, 1010), when(domain.CondRankExactly, 3)}})
		Expect(e.Fit).To(Equal(domain.FitNever))
	})

	It("cannot tell, once all else is met, what the viewer cannot see, and keeps a note even when ready", func() {
		unseen := one(fit(0, 0, "S"), domain.CalendarEntry{Unseen: "Invited", When: []domain.EntryCondition{when(domain.CondRankAtLeast, 5)}})
		Expect(unseen.Fit).To(Equal(domain.FitUnknown))
		Expect(unseen.Need).To(Equal("Invited"))

		short := one(fit(0, 0, "A"), domain.CalendarEntry{Unseen: "Invited", When: []domain.EntryCondition{when(domain.CondRankAtLeast, 5)}})
		Expect(short.Fit).To(Equal(domain.FitNotYet))
		Expect(short.Need).To(Equal("Rank S"))

		noted := one(fit(0, 0, "E"), domain.CalendarEntry{Note: "A clear night"})
		Expect(noted.Fit).To(Equal(domain.FitReady))
		Expect(noted.Need).To(Equal("A clear night"))
	})

	Describe("expeditions", func() {
		pending := func(p *player.Player) { p.Ranch.PhoenixDone = false }
		phoenix := domain.CalendarEntry{Month: 6, Week: 1, Name: domain.Phoenix.Name, Kind: domain.EventTrip,
			When: []domain.EntryCondition{when(domain.CondPhoenixOpen, 0), when(domain.CondYearAtLeast, 1001), when(domain.CondExpedition, 0)}}
		kawrea := domain.CalendarEntry{Month: 6, Week: 1, Name: "Kawrea Volcano", Kind: domain.EventTrip,
			When: []domain.EntryCondition{when(domain.CondPhoenixDone, 0), when(domain.CondExpedition, 0)}}
		both := []domain.CalendarEntry{phoenix, kawrea}

		It("ask what the expedition of that name asks", func() {
			weeks := domain.Calendar(profile(jun1of1001, fit(210, 50, "D"), both))
			Expect(event(on(weeks, 1002, 6, 1), "Kawrea Volcano").Need).To(Equal("Life 280, Rank B"))
		})

		It("open the Phoenix adventure from 1001 while it is not done, and never again once it is", func() {
			open := domain.Calendar(profile(0, fit(140, 40, "D"), both, pending))
			Expect(event(on(open, 1000, 6, 1), domain.Phoenix.Name).Need).To(Equal("Year 1001"))
			Expect(event(on(open, 1001, 6, 1), domain.Phoenix.Name).Fit).To(Equal(domain.FitReady))

			done := domain.Calendar(profile(may4of1001, fit(140, 40, "D"), both))
			Expect(event(on(done, 1001, 6, 1), domain.Phoenix.Name).Fit).To(Equal(domain.FitNever))
		})

		It("keep the other expeditions waiting for the Phoenix adventure", func() {
			waiting := domain.Calendar(profile(may4of1001, fit(300, 60, "B"), both, pending))
			Expect(event(on(waiting, 1001, 6, 1), "Kawrea Volcano").Need).To(Equal("Phoenix first"))
			done := domain.Calendar(profile(may4of1001, fit(300, 60, "B"), both))
			Expect(event(on(done, 1001, 6, 1), "Kawrea Volcano").Fit).To(Equal(domain.FitReady))
		})

		It("cannot tell about the Phoenix adventure where the game has no word for it", func() {
			weeks := domain.Calendar(profile(may4of1001, fit(300, 60, "B"), both, func(p *player.Player) { p.Ranch.PhoenixKnown = false }))
			Expect(event(on(weeks, 1001, 6, 1), domain.Phoenix.Name).Fit).To(Equal(domain.FitUnknown))
			Expect(event(on(weeks, 1001, 6, 1), "Kawrea Volcano").Fit).To(Equal(domain.FitReady))
		})
	})

	It("holds an IMa vs FIMBA entry only in the years of the meet, every 4 years from the first, and none before the game sets it", func() {
		qualifier := domain.CalendarEntry{Month: 7, Week: 4, Name: "Qualifier", When: []domain.EntryCondition{when(domain.CondIMaYear, 0)}}
		set := func(p *player.Player) { p.Ranch.Tournament = 1001 }
		weeks := domain.Calendar(profile(0, fit(0, 0, "D"), []domain.CalendarEntry{qualifier}, set))
		_, in1000 := named(on(weeks, 1000, 7, 4), "Qualifier")
		_, in1001 := named(on(weeks, 1001, 7, 4), "Qualifier")
		Expect(in1000).To(BeFalse())
		Expect(in1001).To(BeTrue())
		later := domain.Calendar(profile(4*48, fit(0, 0, "D"), []domain.CalendarEntry{qualifier}, set))
		_, in1005 := named(on(later, 1005, 7, 4), "Qualifier")
		Expect(in1005).To(BeTrue())

		for _, w := range domain.Calendar(profile(0, fit(0, 0, "D"), []domain.CalendarEntry{qualifier})) {
			_, ok := named(w, "Qualifier")
			Expect(ok).To(BeFalse())
		}
	})

	DescribeTable("offers a facility upgrade when the trainer has the rank its next step needs, and never once there is none",
		func(change func(*player.Player), want domain.Fit, need string) {
			e := one(fit(0, 0, "E"), domain.CalendarEntry{When: []domain.EntryCondition{when(domain.CondFacilityNext, int(player.Barn))}}, change)
			Expect(e.Fit).To(Equal(want))
			Expect(e.Need).To(Equal(need))
		},
		Entry("rank 4", func(p *player.Player) { p.Ranch.Grade = 4 }, domain.FitReady, ""),
		Entry("rank 3", func(p *player.Player) { p.Ranch.Grade = 3 }, domain.FitNotYet, "Trainer rank 4"),
		Entry("no step left", func(p *player.Player) { p.Ranch.Barn, p.Ranch.Grade = 2, 8 }, domain.FitNever, ""),
	)

	It("puts the shop sale as many weeks ahead as the game counts", func() {
		weeks := domain.Calendar(profile(jun1of1001, fit(0, 0, "E"), nil, func(p *player.Player) { p.NextSale = 3 }))
		e := event(weeks[3], domain.ShopSale)
		Expect(e.Kind).To(Equal(domain.EventShop))
		Expect(e.Fit).To(Equal(domain.FitReady))
		for i, w := range weeks {
			_, ok := named(w, domain.ShopSale)
			Expect(ok).To(Equal(i == 3))
		}
	})

	It("says the Errantry sale is on this week when the game says so", func() {
		ready := domain.Calendar(profile(jun1of1001, fit(0, 0, "E"), nil, func(p *player.Player) { p.ErrantryCD = 255 }))
		Expect(event(ready[0], domain.ErrantrySale).Fit).To(Equal(domain.FitReady))
		_, ok := named(domain.Calendar(profile(jun1of1001, fit(0, 0, "E"), nil))[0], domain.ErrantrySale)
		Expect(ok).To(BeFalse())
	})
})
