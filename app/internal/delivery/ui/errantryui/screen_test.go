package errantryui

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	movedomain "app/internal/core/move/domain"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/moveui"
	"app/internal/delivery/ui/profileui"
	mockmove "app/mock/core/move"
	mockprofile "app/mock/core/profile"
)

// whitePixel is the icon of every move, which is all that these tests need of one.
func whitePixel(string) ([]byte, error) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// labelTexts are the words of every label inside an object.
func labelTexts(o fyne.CanvasObject) []string {
	switch o := o.(type) {
	case *widget.Label:
		return []string{o.Text}
	case *fyne.Container:
		var texts []string
		for _, c := range o.Objects {
			texts = append(texts, labelTexts(c)...)
		}
		return texts
	}
	return nil
}

var _ = Describe("Screen", func() {
	var (
		ctx      context.Context
		profiles *mockprofile.MockProfileService
		moves    *mockmove.MockMoveService
		s        *Screen
	)

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		ctx = context.Background()
		profiles = mockprofile.NewMockProfileService(GinkgoT())
		moves = mockmove.NewMockMoveService(GinkgoT())
		s = NewScreen(profiles, moves, moveui.NewIcons(whitePixel))
	})

	It("has a row for each place, naming the place", func() {
		Expect(s.rows).To(HaveLen(5))
		Expect(s.rows[0].place.Text).To(Equal("Papas Mountain"))
		Expect(s.rows[4].place.Text).To(Equal("Kawrea Volcano"))
	})

	It("has no line about the monster above the table, none that says which techs a place teaches, and no note under it", func() {
		s.apply(baby(), grid(), nil)

		for _, text := range labelTexts(s.body) {
			Expect(text).NotTo(ContainSubstring("atigue"))
			Expect(text).NotTo(ContainSubstring("fear"))
			Expect(text).NotTo(ContainSubstring("techs"))
			Expect(text).NotTo(ContainSubstring("The chance that each week"))
		}
	})

	It("is hidden until it has a monster to say anything of", func() {
		Expect(s.body.Visible()).To(BeFalse())
	})

	It("shows the chance of each week for each place", func() {
		s.apply(baby(), grid(), nil)

		Expect(s.body.Visible()).To(BeTrue())
		want := NewView(baby(), grid()).Rows
		for i, row := range s.rows {
			for week := range row.weeks {
				Expect(row.weeks[week].Box()).To(Equal(want[i].Weeks[week]))
			}
		}
		Expect(s.rows[0].weeks[1].Box().Text).To(Equal("69%"))
	})

	It("says what each trip would give, the stat of the place in its colour", func() {
		s.apply(baby(), grid(), nil)

		Expect(s.rows[0].gains.String()).To(Equal("SPD +2-5 · LIF +3"))
		Expect(s.rows[0].gains.Segments()[0].Color).To(Equal(profileui.StatColors[monster.Speed]))
		Expect(s.rows[1].gains.Segments()[0].Color).To(Equal(profileui.StatColors[monster.Pow]))
		Expect(s.rows[4].gains.Segments()[0].Color).To(Equal(profileui.StatColors[monster.Defense]))
	})

	It("puts the icons to the right of the words, under the chances, and not on a line of their own", func() {
		win := test.NewWindow(s.Content())
		win.Resize(fyne.NewSize(480, 900))
		s.apply(baby(), grid(), nil)
		win.Resize(fyne.NewSize(480, 900)) // laid out again, now that there is something to show
		at := func(o fyne.CanvasObject) fyne.Position {
			return fyne.CurrentApp().Driver().AbsolutePositionForObject(o)
		}
		row := s.rows[0]

		Expect(at(row.ready).X).To(BeNumerically(">=", at(row.weeks[0]).X), "in the column of the chances")
		Expect(at(row.ready).Y).To(BeNumerically("<", at(row.gains).Y+row.gains.Size().Height), "not under the words")
		Expect(at(row.ready).Y).To(BeNumerically(">", at(row.weeks[0]).Y), "under the chances")
	})

	It("shows an icon for each move that is ready to learn under the place that teaches it, and says so when there is none", func() {
		s.apply(baby(), grid(), nil)

		Expect(s.rows[0].ready.Objects).To(HaveLen(1), "Ray")
		Expect(s.rows[0].ready.Visible()).To(BeTrue())
		Expect(s.rows[0].none.Visible()).To(BeFalse())
		Expect(s.rows[1].ready.Objects).To(HaveLen(1), "Flame")
		Expect(s.rows[2].ready.Objects).To(BeEmpty())
		Expect(s.rows[2].ready.Visible()).To(BeFalse(), "there is no empty line of icons above the words")
		Expect(s.rows[2].none.Visible()).To(BeTrue())
		Expect(s.rows[2].none.Text).To(Equal("No tech ready"))
	})

	It("writes the chance on the icon, in its bottom right corner, so that it asks for no room of its own", func() {
		win := test.NewWindow(s.Content())
		win.Resize(fyne.NewSize(480, 900))
		s.apply(baby(), grid(), nil)
		win.Resize(fyne.NewSize(480, 900))
		at := func(o fyne.CanvasObject) fyne.Position {
			return fyne.CurrentApp().Driver().AbsolutePositionForObject(o)
		}
		button := s.rows[0].ready.Objects[0].(*iconButton)

		Expect(button.MinSize()).To(Equal(iconSize), "as big as the icon, and no bigger")
		corner := at(button).Add(fyne.NewPos(iconSize.Width, iconSize.Height))
		Expect(at(button.chance).X + button.chance.Size().Width).To(BeNumerically("~", corner.X, 1))
		Expect(at(button.chance).Y + button.chance.Size().Height).To(BeNumerically("~", corner.Y, 1))
	})

	It("gives every place a row as high as the others, whether it has a move to show or not", func() {
		win := test.NewWindow(s.Content())
		win.Resize(fyne.NewSize(480, 900))
		s.apply(baby(), grid(), nil)
		win.Resize(fyne.NewSize(480, 900))

		Expect(s.rows[0].ready.Visible()).To(BeTrue())
		Expect(s.rows[2].noneLine.Size().Height).To(BeNumerically(">=", iconSize.Height), "the line of icons is kept for the words")
		for i, r := range s.rows {
			Expect(r.object.Size().Height).To(BeNumerically("~", s.rows[0].object.Size().Height, 0.5), "row %d", i)
		}
	})

	It("shows no details until the icon of a move is tapped", func() {
		s.apply(baby(), grid(), nil)

		Expect(s.detail.Content().Visible()).To(BeFalse())
	})

	It("shows the details of the move whose icon is tapped, under the table, as the Moves page does, and marks the icon", func() {
		s.apply(baby(), grid(), nil)
		ray, flame := s.rows[0].ready.Objects[0].(*iconButton), s.rows[1].ready.Objects[0].(*iconButton)

		test.Tap(ray)
		Expect(s.detail.Content().Visible()).To(BeTrue())
		Expect(s.detail.Name()).To(Equal("Ray"))
		Expect(ray.selected).To(BeTrue())

		test.Tap(flame)
		Expect(s.detail.Name()).To(Equal("Flame"))
		Expect(ray.selected).To(BeFalse(), "the mark moves to the icon that is tapped")
		Expect(flame.selected).To(BeTrue())
	})

	It("keeps the move that is selected, and its mark, when the page is loaded again", func() {
		s.apply(baby(), grid(), nil)
		test.Tap(s.rows[0].ready.Objects[0].(*iconButton))

		again := grid()
		again.Cells[0][0].Used = 7
		s.apply(baby(), again, nil)

		Expect(s.detail.Content().Visible()).To(BeTrue())
		Expect(s.detail.Name()).To(Equal("Ray"))
		Expect(s.rows[0].ready.Objects[0].(*iconButton).selected).To(BeTrue(), "the icon is a new one, and marked")
	})

	It("takes the details away when the move that is selected is not ready any more", func() {
		s.apply(baby(), grid(), nil)
		test.Tap(s.rows[0].ready.Objects[0].(*iconButton))

		learned := baby()
		learned.Monster.Learned[1] = true // Ray
		s.apply(learned, grid(), nil)

		Expect(s.detail.Content().Visible()).To(BeFalse())
	})

	It("says the chance of each move under its icon", func() {
		s.apply(baby(), grid(), nil)

		Expect(s.rows[0].ready.Objects[0].(*iconButton).chance.Box()).To(Equal(NewView(baby(), grid()).Rows[0].Ready[0].Chance))
		Expect(s.rows[0].ready.Objects[0].(*iconButton).chance.Box().Text).To(Equal("3%"))
	})

	It("takes the icons away when the moves change", func() {
		s.apply(baby(), grid(), nil)
		s.apply(baby(), movedomain.Grid{}, nil)

		Expect(s.rows[0].ready.Objects).To(BeEmpty())
		Expect(s.rows[0].none.Visible()).To(BeTrue())
	})

	It("shows why there is nothing to show, in place of the table", func() {
		s.apply(baby(), grid(), nil)
		s.apply(domain.Profile{}, movedomain.Grid{}, apperr.NotFound("game", "DuckStation is not running"))

		Expect(s.status.Text).To(Equal("DuckStation is not running"))
		Expect(s.status.Visible()).To(BeTrue())
		Expect(s.body.Visible()).To(BeFalse())

		s.apply(baby(), grid(), nil)
		Expect(s.status.Visible()).To(BeFalse())
		Expect(s.body.Visible()).To(BeTrue())
	})

	It("shows any other error as an error", func() {
		s.apply(domain.Profile{}, movedomain.Grid{}, errors.New("db gone"))
		Expect(s.status.Text).To(Equal("Error: db gone"))
	})

	It("asks the services for the monster and for its moves, and shows both", func() {
		profiles.EXPECT().Current(ctx).Return(baby(), nil)
		moves.EXPECT().CurrentGrid(ctx).Return(grid(), nil)
		s.Load(ctx)
		Eventually(func() int { return len(s.rows[0].ready.Objects) }).Should(Equal(1))
		Expect(s.rows[0].weeks[1].Box().Text).To(Equal("69%"))
	})

	It("shows the error of either service", func() {
		profiles.EXPECT().Current(ctx).Return(baby(), nil)
		moves.EXPECT().CurrentGrid(ctx).Return(movedomain.Grid{}, apperr.NotFound("monster", "No monster"))
		s.Load(ctx)
		Eventually(func() string { return s.status.Text }).Should(Equal("No monster"))
	})
})
