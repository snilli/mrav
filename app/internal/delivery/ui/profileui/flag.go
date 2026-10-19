package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Flag is a yes-or-no fact of the monster as a small label: lit in its tone when it is so, dim when it is not.
type Flag struct {
	widget.BaseWidget
	tone Tone
	on   bool
	bg   *canvas.Rectangle
	txt  *canvas.Text
}

func NewFlag(text string, tone Tone) *Flag {
	f := &Flag{
		tone: tone,
		bg:   canvas.NewRectangle(toneNone),
		txt:  canvas.NewText(text, theme.Color(theme.ColorNameDisabled)),
	}
	f.bg.CornerRadius = 4
	f.ExtendBaseWidget(f)
	return f
}

func (f *Flag) On() bool { return f.on }

// SetText changes the words of the flag.
func (f *Flag) SetText(text string) {
	f.txt.Text = text
	f.Refresh()
}

func (f *Flag) Set(on bool) {
	f.on = on
	switch fill, ok := toneColors[f.tone]; {
	case on && ok:
		f.bg.FillColor, f.txt.Color = fill, color.Black
		if f.tone == ToneGray || f.tone == ToneBlack {
			f.txt.Color = color.White
		}
	default:
		f.bg.FillColor, f.txt.Color = toneNone, theme.Color(theme.ColorNameDisabled)
	}
	f.Refresh()
}

func (f *Flag) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(f.bg, container.NewPadded(f.txt)))
}
