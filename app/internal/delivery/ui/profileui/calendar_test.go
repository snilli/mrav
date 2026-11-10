package profileui_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/profileui"
)

var _ = Describe("CalendarView", func() {
	const jun1of1001 = 56
	when := func(t domain.CondType, v int) []domain.EntryCondition {
		return []domain.EntryCondition{{Type: t, Value: v}}
	}
	entries := []domain.CalendarEntry{
		{Month: 6, Week: 1, Name: domain.Phoenix.Name, Kind: domain.EventTrip, When: when(domain.CondPhoenixOpen, 0)},
		{Month: 6, Week: 1, Name: "Kawrea Volcano", Kind: domain.EventTrip, When: when(domain.CondExpedition, 0)},
		{Month: 6, Week: 4, Name: "IMa Official Cup (A)", Kind: domain.EventCup, Purse: 4000, When: when(domain.CondCupGrade, 4)},
		{Month: 6, Week: 4, Name: "IMa Official Cup (B)", Kind: domain.EventCup, Purse: 3000, When: when(domain.CondCupGrade, 3)},
		{Month: 7, Week: 1, Name: "Dragon Tusk invite", Kind: domain.EventInvite, When: when(domain.CondRankExactly, 3)},
		{Month: 7, Week: 2, Name: "Winners' Cup invite", Kind: domain.EventInvite, When: when(domain.CondRankAtLeast, 5)},
		{Month: 9, Week: 1, Name: "Aliens may visit", Kind: domain.EventVisit, Note: "A clear night"},
		{Month: 2, Week: 1, Name: "Durahan invite", Kind: domain.EventInvite, Unseen: "Old Sheath found", When: when(domain.CondRankExactly, 4)},
		{Month: 5, Week: 4, Name: "House upgrade", Kind: domain.EventRanch},
	}
	profile := func() domain.Profile {
		return domain.Profile{
			Monster:  monster.Monster{Rank: 4},
			Calendar: entries,
			Player:   player.Player{NextSale: 1, Ranch: player.Ranch{Known: true, Date: jun1of1001, Grade: 6, PhoenixKnown: true, PhoenixDone: true}},
		}
	}
	find := func(w profileui.CalWeekView, name string) profileui.CalEventView {
		for _, e := range w.Events {
			if e.Name == name {
				return e
			}
		}
		Fail("no " + name)
		return profileui.CalEventView{}
	}
	names := func(w profileui.CalWeekView) []string {
		var out []string
		for _, e := range w.Events {
			out = append(out, e.Name)
		}
		return out
	}
	jun := func(v profileui.CalendarView) profileui.CalMonthView { return v.Years[0].Months[0] }
	all := func() profileui.CalendarShow {
		return profileui.CalendarShow{domain.FitReady: true, domain.FitNotYet: true, domain.FitUnknown: true, domain.FitWarn: true}
	}
	but := func(fit domain.Fit) profileui.CalendarShow {
		show := all()
		show[fit] = false
		return show
	}

	It("groups the weeks by year and by month, this year and the next", func() {
		v := profileui.NewCalendarView(profile(), all())
		Expect(v.Years).To(HaveLen(2))
		Expect(v.Years[0].Title).To(Equal("1001"))
		Expect(v.Years[1].Title).To(Equal("1002"))
		Expect(v.Years[0].Months).To(HaveLen(7))
		Expect(jun(v).Title).To(Equal("Jun"))
		Expect(v.Years[1].Months).To(HaveLen(12))
		for _, m := range v.Years[1].Months {
			Expect(m.Weeks).To(HaveLen(4))
		}
	})

	It("names each week and marks this one", func() {
		w := jun(profileui.NewCalendarView(profile(), all())).Weeks
		Expect(w[0].Title).To(Equal("wk 1"))
		Expect(w[0].Now).To(BeTrue())
		Expect(w[3].Title).To(Equal("wk 4"))
		Expect(w[3].Now).To(BeFalse())
	})

	It("counts the events of a month", func() {
		m := jun(profileui.NewCalendarView(profile(), all()))
		Expect(m.Count).To(Equal(4))
		Expect(names(m.Weeks[1])).To(Equal([]string{domain.ShopSale}))
	})

	It("names each kind of event", func() {
		v := profileui.NewCalendarView(profile(), all())
		Expect(find(jun(v).Weeks[0], "Kawrea Volcano").Type).To(Equal("Expedition"))
		Expect(find(jun(v).Weeks[1], domain.ShopSale).Type).To(Equal("Shop"))
		Expect(find(jun(v).Weeks[3], "IMa Official Cup (A)").Type).To(Equal("Tournament"))
		Expect(find(v.Years[0].Months[1].Weeks[1], "Winners' Cup invite").Type).To(Equal("Invite"))
		Expect(find(v.Years[1].Months[4].Weeks[3], "House upgrade").Type).To(Equal("Ranch"))
		Expect(find(v.Years[0].Months[3].Weeks[0], "Aliens may visit").Type).To(Equal("Visitor"))
	})

	It("tags an event by whether the monster can go, saying what it lacks or what the viewer cannot see", func() {
		v := profileui.NewCalendarView(profile(), all())
		Expect(find(jun(v).Weeks[3], "IMa Official Cup (A)")).To(Equal(profileui.CalEventView{
			Name: "IMa Official Cup (A)", Type: "Tournament", Detail: "4000G",
			Status: profileui.Box{Text: "Ready", Tone: profileui.ToneLimeGreen},
		}))
		Expect(find(v.Years[0].Months[1].Weeks[1], "Winners' Cup invite")).To(Equal(profileui.CalEventView{
			Name: "Winners' Cup invite", Type: "Invite", Note: "needs Rank S",
			Status: profileui.Box{Text: "Not yet", Tone: profileui.ToneOrange},
		}))
		durahan := find(v.Years[1].Months[1].Weeks[0], "Durahan invite")
		Expect(durahan.Note).To(Equal("Old Sheath found?"))
		Expect(durahan.Status).To(Equal(profileui.Box{Text: "Maybe", Tone: profileui.ToneYellow}))
		aliens := find(v.Years[0].Months[3].Weeks[0], "Aliens may visit")
		Expect(aliens.Status.Text).To(Equal("Ready"))
		Expect(aliens.Note).To(Equal("A clear night"))
	})

	It("tags a tournament below the monster's rank as over rank, rather than leaving it out", func() {
		v := profileui.NewCalendarView(profile(), all())
		Expect(find(jun(v).Weeks[3], "IMa Official Cup (B)")).To(Equal(profileui.CalEventView{
			Name: "IMa Official Cup (B)", Type: "Tournament", Detail: "3000G",
			Status: profileui.Box{Text: "Over rank", Tone: profileui.ToneOrangeRed},
		}))
	})

	It("leaves out what will never come for this monster", func() {
		v := profileui.NewCalendarView(profile(), all())
		Expect(names(v.Years[0].Months[1].Weeks[0])).NotTo(ContainElement("Dragon Tusk invite"))
		Expect(names(jun(v).Weeks[0])).NotTo(ContainElement(domain.Phoenix.Name))
	})

	It("hides the cups the monster's rank is past when asked, and counts only what it shows", func() {
		shown := jun(profileui.NewCalendarView(profile(), all()))
		hidden := jun(profileui.NewCalendarView(profile(), but(domain.FitWarn)))
		Expect(names(shown.Weeks[3])).To(ContainElement("IMa Official Cup (B)"))
		Expect(names(hidden.Weeks[3])).NotTo(ContainElement("IMa Official Cup (B)"))
		Expect(names(hidden.Weeks[3])).To(ContainElement("IMa Official Cup (A)"))
		Expect(hidden.Count).To(Equal(shown.Count - 1))
	})

	It("hides what the monster cannot enter yet when asked, and counts only what it shows", func() {
		shown := jun(profileui.NewCalendarView(profile(), all()))
		hidden := jun(profileui.NewCalendarView(profile(), but(domain.FitNotYet)))
		Expect(find(shown.Weeks[0], "Kawrea Volcano").Status.Text).To(Equal("Not yet"))
		Expect(names(hidden.Weeks[0])).NotTo(ContainElement("Kawrea Volcano"))
		Expect(names(hidden.Weeks[3])).To(ContainElement("IMa Official Cup (B)"))
		Expect(hidden.Count).To(Equal(shown.Count - 1))
	})

	It("hides what the viewer cannot see it all of when asked", func() {
		hidden := profileui.NewCalendarView(profile(), but(domain.FitUnknown))
		Expect(names(hidden.Years[1].Months[1].Weeks[0])).NotTo(ContainElement("Durahan invite"))
	})

	It("counts every event of each status, whether it is shown or not", func() {
		shown := profileui.NewCalendarView(profile(), all())
		hidden := profileui.NewCalendarView(profile(), profileui.CalendarShow{domain.FitReady: true})
		Expect(hidden.Fits).To(Equal(shown.Fits))
		Expect(shown.Fits[domain.FitWarn]).To(BeNumerically(">", 0))
		Expect(shown.Fits).NotTo(HaveKey(domain.FitNever))
	})

	It("explains every status with its colour, in a few words", func() {
		Expect(profileui.CalendarLegend).To(Equal([]profileui.LegendItem{
			{Fit: domain.FitReady, Status: profileui.Box{Text: "Ready", Tone: profileui.ToneLimeGreen}},
			{Fit: domain.FitNotYet, Status: profileui.Box{Text: "Not yet", Tone: profileui.ToneOrange}},
			{Fit: domain.FitUnknown, Status: profileui.Box{Text: "Maybe", Tone: profileui.ToneYellow}},
			{Fit: domain.FitWarn, Status: profileui.Box{Text: "Over rank", Tone: profileui.ToneOrangeRed}},
		}))
	})

	It("says so when the game does not say the date", func() {
		v := profileui.NewCalendarView(domain.Profile{}, all())
		Expect(v.Years).To(BeEmpty())
		Expect(v.None).To(Equal("The game does not say the date yet"))
	})
})
