package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Flag struct {
	widget.BaseWidget
	tone     Tone
	on       bool
	bg       *canvas.Rectangle
	txt      *canvas.Text
	OnTapped func()
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

func (f *Flag) Tapped(*fyne.PointEvent) {
	if f.OnTapped != nil {
		f.OnTapped()
	}
}

func (f *Flag) Cursor() desktop.Cursor {
	if f.OnTapped != nil {
		return desktop.PointerCursor
	}
	return desktop.DefaultCursor
}

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
	f.bg.StrokeColor, f.bg.StrokeWidth = theme.Color(theme.ColorNameDisabled), 0
	if !on && f.OnTapped != nil {
		f.bg.StrokeWidth = 1
	}
	f.Refresh()
}

func (f *Flag) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(container.NewStack(f.bg, container.NewPadded(f.txt)))
}
