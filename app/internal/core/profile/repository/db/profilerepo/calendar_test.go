package profilerepo_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"orm"
	entcalendar "orm/ent/calendarevent"
	"orm/stored"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
	"app/internal/core/profile/repository/db/profilerepo"
	"app/internal/testutil"
)

var _ = Describe("profilerepo calendar", Label("integration"), func() {
	var (
		ctx    context.Context
		client *orm.Client
	)
	min := func(n int) *int { return &n }

	BeforeEach(func() {
		ctx = context.Background()
		var err error
		client, err = testutil.NewClient(ctx)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(client.Close)
	})

	It("lists the calendar entries of a game in the order of the year, with every field and condition", func() {
		client.CalendarEvent.Create().SetVersion(entcalendar.VersionMr2dx).SetMonth(2).SetWeek(1).SetName("Durahan invite").
			SetKind(entcalendar.KindInvite).SetUnseen("Old Sheath found").
			SetConditions([]stored.Condition{{Type: stored.RankExactly, Min: min(4)}}).SaveX(ctx)
		client.CalendarEvent.Create().SetVersion(entcalendar.VersionMr2dx).SetMonth(1).SetWeek(2).SetName("Sirius Cup (B)").
			SetKind(entcalendar.KindCup).SetPurse(4000).SetPrize("Disc Chips (Ape)").SetNote("A note").SaveX(ctx)
		client.CalendarEvent.Create().SetVersion(entcalendar.VersionMr2).SetMonth(1).SetWeek(1).SetName("PlayStation only").
			SetKind(entcalendar.KindVisit).SaveX(ctx)

		got, err := profilerepo.New(client).ListCalendar(ctx, monster.MR2DX)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]domain.CalendarEntry{
			{Month: 1, Week: 2, Name: "Sirius Cup (B)", Kind: domain.EventCup, Purse: 4000, Prize: "Disc Chips (Ape)", Note: "A note"},
			{Month: 2, Week: 1, Name: "Durahan invite", Kind: domain.EventInvite, Unseen: "Old Sheath found",
				When: []domain.EntryCondition{{Type: domain.CondRankExactly, Value: 4}}},
		}))
	})

	It("reads every kind", func() {
		for i, k := range []entcalendar.Kind{entcalendar.KindCup, entcalendar.KindInvite, entcalendar.KindTrip, entcalendar.KindRanch, entcalendar.KindVisit} {
			client.CalendarEvent.Create().SetVersion(entcalendar.VersionMr2).SetMonth(i + 1).SetWeek(1).SetName(string(k)).SetKind(k).SaveX(ctx)
		}
		got, err := profilerepo.New(client).ListCalendar(ctx, monster.MR2)
		Expect(err).NotTo(HaveOccurred())
		var kinds []domain.EventKind
		for _, e := range got {
			kinds = append(kinds, e.Kind)
		}
		Expect(kinds).To(Equal([]domain.EventKind{domain.EventCup, domain.EventInvite, domain.EventTrip, domain.EventRanch, domain.EventVisit}))
	})

	It("fails, saying which entry, for a condition it does not know", func() {
		client.CalendarEvent.Create().SetVersion(entcalendar.VersionMr2).SetMonth(1).SetWeek(1).SetName("Odd").SetKind(entcalendar.KindCup).
			SetConditions([]stored.Condition{{Type: "moon_phase"}}).SaveX(ctx)
		_, err := profilerepo.New(client).ListCalendar(ctx, monster.MR2)
		Expect(err).To(MatchError(And(ContainSubstring("Odd"), ContainSubstring("moon_phase"))))
	})

	Describe("on the shipped database", func() {
		var shipped *orm.Client
		BeforeEach(func() {
			var err error
			shipped, err = testutil.OpenShipped(ctx)
			if errors.Is(err, testutil.ErrNoShippedDB) {
				Skip(err.Error())
			}
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(shipped.Close)
		})
		entry := func(entries []domain.CalendarEntry, name string) domain.CalendarEntry {
			for _, e := range entries {
				if e.Name == name {
					return e
				}
			}
			Fail("no " + name)
			return domain.CalendarEntry{}
		}

		It("has the same calendar for both games", func() {
			repo := profilerepo.New(shipped)
			mr2, err := repo.ListCalendar(ctx, monster.MR2)
			Expect(err).NotTo(HaveOccurred())
			mr2dx, err := repo.ListCalendar(ctx, monster.MR2DX)
			Expect(err).NotTo(HaveOccurred())
			Expect(mr2dx).To(HaveLen(89))
			Expect(mr2).To(Equal(mr2dx))
		})

		It("holds the tournaments as LegendCup lists them", func() {
			got, err := profilerepo.New(shipped).ListCalendar(ctx, monster.MR2DX)
			Expect(err).NotTo(HaveOccurred())
			cups := 0
			for _, e := range got {
				if e.Kind == domain.EventCup {
					cups++
				}
			}
			Expect(cups).To(Equal(49 + 24))
			Expect(entry(got, "Poannka Cup (B)")).To(Equal(domain.CalendarEntry{
				Month: 2, Week: 2, Name: "Poannka Cup (B)", Kind: domain.EventCup, Purse: 5000, Prize: "Disc Chips (Hopper)",
				When: []domain.EntryCondition{{Type: domain.CondCupGrade, Value: 3}},
			}))
			Expect(entry(got, "Invitational (A)").When).To(Equal([]domain.EntryCondition{{Type: domain.CondRankExactly, Value: 4}}))
			for _, letter := range []string{"S", "A", "B", "C", "D", "E"} {
				Expect(entry(got, "IMa Official Cup ("+letter+")").Month).To(Equal(3))
			}
			Expect(entry(got, "New Year Cup (F)").When).To(BeEmpty())
			Expect(entry(got, "Elders' Cup (F)").When).To(Equal([]domain.EntryCondition{{Type: domain.CondAgeAtLeast, Value: 300}}))
			Expect(entry(got, "Greatest 4 (S)").Unseen).To(Equal("Invited"))
		})

		It("holds the invitations, expeditions, upgrades and visits", func() {
			got, err := profilerepo.New(shipped).ListCalendar(ctx, monster.MR2DX)
			Expect(err).NotTo(HaveOccurred())
			Expect(entry(got, "Durahan invite")).To(Equal(domain.CalendarEntry{
				Month: 2, Week: 1, Name: "Durahan invite", Kind: domain.EventInvite, Unseen: "Old Sheath found",
				When: []domain.EntryCondition{{Type: domain.CondRankExactly, Value: 4}},
			}))
			Expect(entry(got, domain.Phoenix.Name).When).To(Equal([]domain.EntryCondition{
				{Type: domain.CondPhoenixOpen}, {Type: domain.CondYearAtLeast, Value: 1001}, {Type: domain.CondExpedition},
			}))
			for _, e := range domain.Expeditions {
				Expect(entry(got, e.Name).Month).To(Equal(e.Month))
			}
			Expect(entry(got, "Aliens may visit").Kind).To(Equal(domain.EventVisit))
			Expect(entry(got, "IMa vs FIMBA qualifier").When).To(ContainElement(domain.EntryCondition{Type: domain.CondIMaYear}))
		})
	})
})
