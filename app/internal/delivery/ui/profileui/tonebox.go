package profileui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// toneNone is no fill at all, so that the theme shows through.
var toneNone color.Color = color.Transparent

// toneColors are the fills of the tones, which are MR2AV's colours, and the blue of the bars.
var toneColors = map[Tone]color.Color{
	ToneRed:         color.NRGBA{R: 255, A: 255},
	ToneOrangeRed:   color.NRGBA{R: 255, G: 69, A: 255},
	ToneOrange:      color.NRGBA{R: 255, G: 165, A: 255},
	ToneYellow:      color.NRGBA{R: 255, G: 255, A: 255},
	ToneYellowGreen: color.NRGBA{R: 154, G: 205, B: 50, A: 255},
	ToneLimeGreen:   color.NRGBA{R: 50, G: 205, B: 50, A: 255},
	ToneLightGreen:  color.NRGBA{R: 144, G: 238, B: 144, A: 255},
	ToneLightYellow: color.NRGBA{R: 255, G: 255, B: 224, A: 255},
	ToneLightPink:   color.NRGBA{R: 255, G: 182, B: 193, A: 255},
	ToneHotPink:     color.NRGBA{R: 255, G: 105, B: 180, A: 255},
	ToneGray:        color.NRGBA{R: 169, G: 169, B: 169, A: 255},
	ToneBlack:       color.NRGBA{A: 255},
	ToneAccent:      color.NRGBA{R: 91, G: 141, B: 239, A: 255},
}

// ToneBox shows a value on the colour of its tone: black text, or white on the two dark fills.
type ToneBox struct {
	widget.BaseWidget
	bg      *canvas.Rectangle
	txt     *canvas.Text
	box     Box
	compact bool
}

func NewToneBox() *ToneBox {
	b := &ToneBox{
		bg:  canvas.NewRectangle(toneNone),
		txt: canvas.NewText("", theme.Color(theme.ColorNameForeground)),
	}
	b.bg.CornerRadius = 4
	b.ExtendBaseWidget(b)
	return b
}

// Large makes the box's text big and bold, for a title.
func (b *ToneBox) Large() *ToneBox {
	b.txt.TextSize = 22
	b.txt.TextStyle = fyne.TextStyle{Bold: true}
	return b
}

// Compact makes the box small, for the corner of an icon: small bold text with hardly any room round it.
func (b *ToneBox) Compact() *ToneBox {
	b.txt.TextSize = 10
	b.txt.TextStyle = fyne.TextStyle{Bold: true}
	b.compact = true
	return b
}

// Box is what the box was last given to show.
func (b *ToneBox) Box() Box { return b.box }

func (b *ToneBox) Set(box Box) {
	b.box = box
	b.txt.Text = box.Text
	if fill, ok := toneColors[box.Tone]; ok {
		b.bg.FillColor, b.txt.Color = fill, color.Black
		if box.Tone == ToneGray || box.Tone == ToneBlack {
			b.txt.Color = color.White
		}
	} else {
		b.bg.FillColor, b.txt.Color = toneNone, theme.Color(theme.ColorNameForeground)
	}
	b.Refresh()
}

func (b *ToneBox) CreateRenderer() fyne.WidgetRenderer {
	padded := fyne.CanvasObject(container.NewPadded(b.txt))
	if b.compact {
		padded = container.New(layout.NewCustomPaddedLayout(0, 0, 2, 2), b.txt)
	}
	return widget.NewSimpleRenderer(container.NewStack(b.bg, padded))
}

func setVisible(o fyne.CanvasObject, on bool) {
	if on == o.Visible() {
		return
	}
	if on {
		o.Show()
	} else {
		o.Hide()
	}
}
