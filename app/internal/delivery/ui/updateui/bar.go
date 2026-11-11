package updateui

import (
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/rs/zerolog/log"

	"app/internal/core/release/domain"
)

type Bar struct {
	open     func(*url.URL) error
	release  domain.Release
	message  *widget.Label
	download *widget.Button
	close    *widget.Button
	row      *fyne.Container
	object   *fyne.Container
}

func New(open func(*url.URL) error) *Bar {
	b := &Bar{open: open, message: widget.NewLabel("")}
	b.download = widget.NewButton("Download", b.openPage)
	b.download.Importance = widget.HighImportance
	b.close = widget.NewButtonWithIcon("", theme.CancelIcon(), b.hide)
	b.close.Importance = widget.LowImportance
	b.row = container.NewBorder(nil, nil, nil, container.NewHBox(b.download, b.close), b.message)
	b.row.Hide()
	b.object = container.NewStack(b.row)
	return b
}

func (b *Bar) Object() fyne.CanvasObject { return b.object }

func (b *Bar) Offer(r domain.Release) {
	fyne.Do(func() { b.show(r) })
}

func (b *Bar) show(r domain.Release) {
	b.release = r
	b.message.SetText("mrav " + r.Tag + " is out.")
	b.row.Show()
	b.object.Refresh()
}

func (b *Bar) hide() {
	b.row.Hide()
	b.object.Refresh()
}

func (b *Bar) openPage() {
	u, err := url.Parse(b.release.URL)
	if err == nil {
		err = b.open(u)
	}
	if err != nil {
		log.Error().Err(err).Str("url", b.release.URL).Msg("cannot open the release page")
	}
}
