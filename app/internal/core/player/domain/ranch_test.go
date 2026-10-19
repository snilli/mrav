package domain_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/player/domain"
)

var _ = Describe("Ranch", func() {
	at := func(weeks int) domain.Ranch { return domain.Ranch{Known: true, Date: weeks} }

	DescribeTable("reads the date as the game counts it: from April of the year 1000, four weeks to a month, twelve to a year",
		func(weeks int, want domain.Date) { Expect(at(weeks).Calendar()).To(Equal(want)) },
		Entry("the first week of the game is April's", 0, domain.Date{Year: 1000, Month: 4, Week: 1}),
		Entry("four weeks on is May's", 4, domain.Date{Year: 1000, Month: 5, Week: 1}),
		Entry("the last week of the first year", 35, domain.Date{Year: 1000, Month: 12, Week: 4}),
		Entry("and the first of the next", 36, domain.Date{Year: 1001, Month: 1, Week: 1}),
		Entry("the last of that one", 83, domain.Date{Year: 1001, Month: 12, Week: 4}),
		Entry("week 860 is March, week 1, of 1018: the reading that this was worked out from", 860, domain.Date{Year: 1018, Month: 3, Week: 1}),
	)

	It("writes a date as the game would: the year, the month and the week", func() {
		Expect(domain.Date{Year: 1002, Month: 5, Week: 4}.String()).To(Equal("Y1002 May wk 4"))
		Expect(domain.Date{Year: 1000, Month: 1, Week: 1}.String()).To(Equal("Y1000 Jan wk 1"))
	})

	DescribeTable("counts the weeks until the calendar next reaches a week of the year, now included",
		func(weeks, want int) {
			Expect(at(weeks).WeeksUntil(domain.UpgradeMonth, domain.UpgradeWeek)).To(Equal(want))
		},
		Entry("from the very start, in April, May week 4 is 7 weeks on", 0, 7),
		Entry("on the day itself it is 0", 7, 0),
		Entry("the week after, it is next year's", 8, 47),
		Entry("in any year", 48+7, 0),
		Entry("from the end of the year", 35, 20),
	)

	DescribeTable("finds the next IMA-FIMBA tournament: July, week 1, every 4 years from the year of the first",
		func(weeks int, want domain.Date, wantWeeks int) {
			r := at(weeks)
			r.Tournament = 1003
			when, away, ok := r.NextTournament()
			Expect(ok).To(BeTrue())
			Expect(when).To(Equal(want))
			Expect(away).To(Equal(wantWeeks))
		},
		Entry("before the first one", 0, domain.Date{Year: 1003, Month: 7, Week: 1}, 156),
		Entry("in March of the year after one, the next is 64 weeks away", 860, domain.Date{Year: 1019, Month: 7, Week: 1}, 64),
		Entry("in March of the year of one, it is this year's", 908, domain.Date{Year: 1019, Month: 7, Week: 1}, 16),
		Entry("on the day itself it is 0", 924, domain.Date{Year: 1019, Month: 7, Week: 1}, 0),
		Entry("the week after it, it is four years on, less the week", 925, domain.Date{Year: 1023, Month: 7, Week: 1}, 191),
	)

	It("does not know when the tournament is where the game's year of the first was not read", func() {
		_, _, ok := at(860).NextTournament()
		Expect(ok).To(BeFalse())
		_, _, ok = domain.Ranch{Date: 860, Tournament: 1003}.NextTournament()
		Expect(ok).To(BeFalse(), "a ranch that is not known has no date")
	})

	DescribeTable("reads a tile of an expedition's map as its column and row: 32 to a row, the row first",
		func(tile, x, y int) {
			gotX, gotY := domain.Ranch{Tile: tile}.Spot()
			Expect([]int{gotX, gotY}).To(Equal([]int{x, y}))
		},
		Entry("column 15 of row 30", 975, 15, 30),
		Entry("one step up is a row less", 943, 15, 29),
		Entry("two steps up", 911, 15, 28),
		Entry("one step right is a column more", 912, 16, 28),
		Entry("the corner", 0, 0, 0),
	)

	DescribeTable("says what a facility's next upgrade is, or that there is none",
		func(r domain.Ranch, f domain.Facility, want domain.Upgrade, ok bool) {
			got, found := r.Next(f)
			Expect(found).To(Equal(ok))
			Expect(got).To(Equal(want))
		},
		Entry("a house as it starts", domain.Ranch{House: 0}, domain.House, domain.Upgrade{Cost: 5000, Grade: 3}, true),
		Entry("a house one up", domain.Ranch{House: 1}, domain.House, domain.Upgrade{Cost: 8000, Grade: 5}, true),
		Entry("a house three up", domain.Ranch{House: 3}, domain.House, domain.Upgrade{Cost: 20000, Grade: 8}, true),
		Entry("a house at its top", domain.Ranch{House: 4}, domain.House, domain.Upgrade{}, false),
		Entry("a barn as it starts", domain.Ranch{Barn: 0}, domain.Barn, domain.Upgrade{Cost: 15000, Grade: 4}, true),
		Entry("a barn one up", domain.Ranch{Barn: 1}, domain.Barn, domain.Upgrade{Cost: 30000, Grade: 7}, true),
		Entry("a barn at its top", domain.Ranch{Barn: 2}, domain.Barn, domain.Upgrade{}, false),
		Entry("no altar has none to upgrade", domain.Ranch{Altar: 0}, domain.Altar, domain.Upgrade{}, false),
		Entry("the altar it starts with", domain.Ranch{Altar: 1}, domain.Altar, domain.Upgrade{Cost: 5000, Needed: 8000}, true),
		Entry("the altar one up", domain.Ranch{Altar: 2}, domain.Altar, domain.Upgrade{Cost: 8000, Needed: 13000}, true),
		Entry("an altar at its top", domain.Ranch{Altar: 3}, domain.Altar, domain.Upgrade{}, false),
	)

	It("adds up to 90,000 G for every upgrade of the house and the barn, as the guide says", func() {
		total := 0
		for _, f := range []domain.Facility{domain.House, domain.Barn} {
			for level := range domain.MaxLevel(f) {
				r := domain.Ranch{House: level, Barn: level}
				up, ok := r.Next(f)
				Expect(ok).To(BeTrue())
				total += up.Cost
			}
		}
		Expect(total).To(Equal(90000))
	})

	It("needs the gold of the upgrade and 5,000 more on hand for the game to offer it", func() {
		Expect(domain.Upgrade{Cost: 8000}.GoldNeeded()).To(Equal(13000))
	})

	It("needs the gold that the guide gives for an upgrade that it gives one for, as the altar's does", func() {
		Expect(domain.Upgrade{Cost: 5000, Needed: 8000}.GoldNeeded()).To(Equal(8000))
	})

	It("offers the house and the stable upgrades in May week 4, and the altar's in March week 3, 140 weeks apart", func() {
		for _, f := range []domain.Facility{domain.House, domain.Barn} {
			month, week := domain.UpgradeTime(f)
			Expect([]int{month, week}).To(Equal([]int{5, 4}))
		}
		month, week := domain.UpgradeTime(domain.Altar)
		Expect([]int{month, week}).To(Equal([]int{3, 3}))
		Expect(domain.AltarWait).To(Equal(140))
	})

	It("knows how far each facility can go", func() {
		Expect(domain.MaxLevel(domain.House)).To(Equal(4))
		Expect(domain.MaxLevel(domain.Barn)).To(Equal(2))
		Expect(domain.MaxLevel(domain.Altar)).To(Equal(3))
	})

	DescribeTable("says how many places for items the house gives: 12, and 2 more for each upgrade",
		func(r domain.Ranch, want int) { Expect(r.Slots()).To(Equal(want)) },
		Entry("a house as it starts", domain.Ranch{Known: true, House: 0}, 12),
		Entry("one up", domain.Ranch{Known: true, House: 1}, 14),
		Entry("fully upgraded", domain.Ranch{Known: true, House: 4}, 20),
		Entry("never more than the game has", domain.Ranch{Known: true, House: 9}, 20),
		Entry("all of them when the house is not known", domain.Ranch{House: 0}, 20),
	)
})
