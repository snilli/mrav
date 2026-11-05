package profileui

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
)

var _ = Describe("CalendarPage", func() {
	BeforeEach(func() { test.NewTempApp(GinkgoTB()) })

	heads := func(c *CalendarPage) []*widget.Button {
		var out []*widget.Button
		var walk func(o fyne.CanvasObject)
		walk = func(o fyne.CanvasObject) {
			switch o := o.(type) {
			case *widget.Button:
				out = append(out, o)
			case *fyne.Container:
				for _, child := range o.Objects {
					walk(child)
				}
			}
		}
		walk(c.years)
		return out
	}

	It("folds a year away and back with its header, keeping it so when the data comes again", func() {
		c := NewCalendarPage()
		p := domain.Profile{Player: player.Player{NextSale: 99, Ranch: player.Ranch{Known: true, Date: 56}}}
		c.Set(&p)
		year := c.years.Objects[0].(*fyne.Container)
		body := year.Objects[1]
		Expect(heads(c)[0].Text).To(Equal("▾ 1001"))
		Expect(body.Visible()).To(BeTrue())

		test.Tap(heads(c)[0])
		Expect(heads(c)[0].Text).To(Equal("▸ 1001"))
		Expect(body.Visible()).To(BeFalse())

		p.Player.NextSale = 3
		c.Set(&p)
		Expect(heads(c)[0].Text).To(Equal("▸ 1001"))
		Expect(c.years.Objects[0].(*fyne.Container).Objects[1].Visible()).To(BeFalse())

		test.Tap(heads(c)[0])
		Expect(c.years.Objects[0].(*fyne.Container).Objects[1].Visible()).To(BeTrue())
	})

	labels := func(o fyne.CanvasObject) []*widget.Label {
		var out []*widget.Label
		var walk func(o fyne.CanvasObject)
		walk = func(o fyne.CanvasObject) {
			switch o := o.(type) {
			case *widget.Label:
				out = append(out, o)
			case *fyne.Container:
				for _, child := range o.Objects {
					walk(child)
				}
			}
		}
		walk(o)
		return out
	}
	texts := func(o fyne.CanvasObject) []string {
		var out []string
		for _, l := range labels(o) {
			out = append(out, l.Text)
		}
		return out
	}

	It("names every event the same way: its kind in a quiet line above, then its name alone", func() {
		for _, e := range []CalEventView{{Name: "Sirius Cup (B)", Type: "Tournament"}, {Name: "Shop sale", Type: "Shop"}, {Name: "Durahan invite", Type: "Invite"}} {
			got := labels(eventRow(e))
			Expect(got[0].Text).To(Equal(e.Type))
			Expect(got[0].Importance).To(Equal(widget.LowImportance))
			Expect(got[1].Text).To(Equal(e.Name))
			Expect(got[1].TextStyle.Bold).To(BeTrue())
		}
	})

	It("marks this week by its whole row, with now in its title, and no other week", func() {
		now, nowFill := weekRow(CalWeekView{Title: "wk 1", Now: true})
		other, otherFill := weekRow(CalWeekView{Title: "wk 2"})
		Expect(texts(now)).To(ContainElement("wk 1 · now"))
		Expect(texts(now)).NotTo(ContainElement("now"))
		Expect(texts(other)).To(ContainElement("wk 2"))
		Expect(nowFill.Visible()).To(BeTrue())
		Expect(otherFill.Visible()).To(BeFalse())
	})

	It("has a box to hide what the monster cannot enter yet, beside the one for cups past its rank", func() {
		c := NewCalendarPage()
		Expect(c.hideNotYet.Text).To(Equal("Hide not yet"))
		Expect(c.hideNotYet.Checked).To(BeFalse())
	})

	It("keeps what the colours mean folded above the calendar, to open with one tap, and keeps it open when the data comes again", func() {
		c := NewCalendarPage()
		p := domain.Profile{Player: player.Player{NextSale: 99, Ranch: player.Ranch{Known: true, Date: 56}}}
		c.Set(&p)
		page := c.content.Content.(*fyne.Container).Objects
		Expect(slices.Index(page, c.legend)).To(BeNumerically("<", slices.Index(page, fyne.CanvasObject(c.years))))
		head := c.legend.(*fyne.Container).Objects[0].(*widget.Button)
		body := c.legend.(*fyne.Container).Objects[1]
		Expect(head.Text).To(Equal("▸ What the colours mean"))
		Expect(body.Visible()).To(BeFalse())

		test.Tap(head)
		Expect(head.Text).To(Equal("▾ What the colours mean"))
		Expect(body.Visible()).To(BeTrue())
		p.Player.NextSale = 3
		c.Set(&p)
		Expect(body.Visible()).To(BeTrue())
	})

	It("asks for no more width than the Vitals page, so a column of it is no wider than the others", func() {
		c := NewCalendarPage()
		c.Set(&domain.Profile{Player: player.Player{NextSale: 1, Ranch: player.Ranch{Known: true, Date: 49}}})
		Expect(c.content.Content.MinSize().Width).To(BeNumerically("<=", NewDashboard(nil).vitalsBody.MinSize().Width))
	})

	It("opens this month and the next, and keeps the others folded", func() {
		c := NewCalendarPage()
		c.Set(&domain.Profile{Player: player.Player{NextSale: 99, Ranch: player.Ranch{Known: true, Date: 56}}})
		var months []string
		for _, b := range heads(c)[1:4] {
			months = append(months, b.Text)
		}
		Expect(months).To(Equal([]string{"▾ Jun (0)", "▾ Jul (0)", "▸ Aug (0)"}))
	})

	It("opens the new month and the next when the date moves on, and keeps a head the viewer folded", func() {
		c := NewCalendarPage()
		p := domain.Profile{Player: player.Player{NextSale: 99, Ranch: player.Ranch{Known: true, Date: 0}}}
		c.Set(&p)
		p.Player.Ranch.Date = 8
		c.Set(&p)
		Expect(heads(c)[0].Text).To(HavePrefix("▾ "))
		Expect(heads(c)[1].Text).To(HavePrefix("▾ "))
		Expect(heads(c)[2].Text).To(HavePrefix("▾ "))

		p.Player.Ranch.Date = 44
		c.Set(&p)
		Expect(heads(c)[0].Text).To(HavePrefix("▾ "))
		Expect(heads(c)[1].Text).To(HavePrefix("▾ "))

		test.Tap(heads(c)[1])
		p.Player.NextSale = 3
		c.Set(&p)
		Expect(heads(c)[1].Text).To(HavePrefix("▸ "))
	})
})
