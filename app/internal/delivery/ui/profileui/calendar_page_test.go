package profileui

import (
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

	It("names a tournament or a sale on its own, and puts the kind before any other event", func() {
		Expect(eventTitle(CalEventView{Name: "Sirius Cup (B)", Type: "Tournament"})).To(Equal("Sirius Cup (B)"))
		Expect(eventTitle(CalEventView{Name: "Shop sale", Type: "Shop"})).To(Equal("Shop sale"))
		Expect(eventTitle(CalEventView{Name: "Durahan invite", Type: "Invite"})).To(Equal("Invite · Durahan invite"))
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
