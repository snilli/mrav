package itemui

import (
	"context"
	"errors"
	"image/png"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/software"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	"app/internal/core/item/domain"
	player "app/internal/core/player/domain"
	mockitem "app/mock/core/item"
)

// slots are the twenty places, with a Magic Banana in the second and the rest empty.
func slots() []domain.Slot {
	out := make([]domain.Slot, player.ItemSlots)
	for i := range out {
		out[i].Index = i
	}
	out[1].Item = &domain.Item{
		Code: 29, Name: "Magic Banana", Kind: "Mysterious Banana", Description: "Its effect differs each time.",
		Effect: "[effect]", SalePrice: "100", PurchasePrice: "200",
	}
	return out
}

var _ = Describe("Screen", func() {
	var (
		ctx context.Context
		svc *mockitem.MockItemService
		s   *Screen
	)

	BeforeEach(func() {
		test.NewTempApp(GinkgoTB())
		ctx = context.Background()
		svc = mockitem.NewMockItemService(GinkgoT())
		s = NewScreen(svc)
	})

	It("has a button for each of the twenty places, and shows nothing selected", func() {
		Expect(s.buttons).To(HaveLen(player.ItemSlots))
		Expect(s.description.Text).To(Equal("No item selected."))
		Expect(s.effect.Text).To(Equal("No item selected."))
	})

	Describe("the twenty buttons", func() {
		// rows is how many lines the buttons fall on when the page is as wide as asked.
		rows := func(width float32) int {
			win := test.NewWindow(s.Content())
			win.Resize(fyne.NewSize(width, 900))
			ys := map[float32]bool{}
			for _, b := range s.buttons {
				ys[fyne.CurrentApp().Driver().AbsolutePositionForObject(b).Y] = true
			}
			return len(ys)
		}

		It("are in two rows when the page is wide, as the game lays them out", func() {
			Expect(rows(1500)).To(Equal(2))
		})

		It("wrap onto more rows when the page is narrow, and never make the page wider", func() {
			Expect(rows(400)).To(BeNumerically(">", 4))
			Expect(s.strip.MinSize().Width).To(BeNumerically("<", 200))
		})
	})

	It("labels each button with the item in its place, and dashes for an empty place", func() {
		s.apply(slots(), nil)
		Expect(s.buttons[1].Text).To(Equal("Magic Banana"))
		Expect(s.buttons[0].Text).To(Equal("-----"))
		Expect(s.buttons[19].Text).To(Equal("-----"))
	})

	It("shows what the list says of the item of a button that is tapped", func() {
		s.apply(slots(), nil)
		test.Tap(s.buttons[1])
		Expect(s.name.Text).To(Equal("Magic Banana"))
		Expect(s.kind.Text).To(Equal("Mysterious Banana"))
		Expect(s.description.Text).To(Equal("Its effect differs each time."))
		Expect(s.effect.Text).To(Equal("[effect]"))
		Expect(s.sale.Text).To(Equal("100"))
		Expect(s.purchase.Text).To(Equal("200"))
	})

	It("dims the places the house has no room for, and takes the dimming away when it is upgraded", func() {
		locked := slots()
		for i := 14; i < len(locked); i++ {
			locked[i].Locked = true
		}
		s.apply(locked, nil)
		Expect(s.buttons[13].Disabled()).To(BeFalse())
		Expect(s.buttons[14].Disabled()).To(BeTrue())
		Expect(s.buttons[14].Text).To(Equal("locked"))

		s.apply(slots(), nil)
		Expect(s.buttons[14].Disabled()).To(BeFalse())
		Expect(s.buttons[14].Text).To(Equal("-----"))
	})

	It("shows a blank place as one", func() {
		s.apply(slots(), nil)
		test.Tap(s.buttons[0])
		Expect(s.kind.Text).To(Equal("No Item"))
		Expect(s.effect.Text).To(Equal("This is a blank slot."))
	})

	It("keeps the place selected, and brings its details up to date, when the slots change", func() {
		s.apply(slots(), nil)
		test.Tap(s.buttons[0])
		Expect(s.effect.Text).To(Equal("This is a blank slot."))

		changed := slots()
		changed[0].Item = changed[1].Item
		s.apply(changed, nil)
		Expect(s.name.Text).To(Equal("Magic Banana"))
	})

	It("shows why there is nothing to show, in place of the page", func() {
		s.apply(nil, apperr.NotFound("game", "MR2DX is not running"))
		Expect(s.status.Text).To(Equal("MR2DX is not running"))
		Expect(s.body.Visible()).To(BeFalse())

		s.apply(slots(), nil)
		Expect(s.body.Visible()).To(BeTrue())
	})

	It("shows any other error as an error", func() {
		s.apply(nil, errors.New("db gone"))
		Expect(s.status.Text).To(Equal("Error: db gone"))
	})

	It("asks the service for the inventory and shows it", func() {
		svc.EXPECT().Inventory(ctx).Return(slots(), nil)
		s.Load(ctx)
		Eventually(func() string { return s.buttons[1].Text }).Should(Equal("Magic Banana"))
	})

	// A screenshot of the page, for looking at: SCREENSHOT_DIR=/some/dir go test -ginkgo.label-filter=screenshot
	It("renders the page to a PNG", Label("screenshot"), func() {
		dir := os.Getenv("SCREENSHOT_DIR")
		if dir == "" {
			Skip("set SCREENSHOT_DIR to write the screenshots")
		}
		s.apply(slots(), nil)
		test.Tap(s.buttons[1])
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		f, err := os.Create(filepath.Join(dir, "items-page.png"))
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(f.Close)
		Expect(png.Encode(f, software.Render(container.NewVBox(s.status, s.body), theme.DefaultTheme()))).To(Succeed())
	})
})
