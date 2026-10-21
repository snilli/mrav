package shellui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Theme is the default theme, always dark whatever the system asks for, with less padding around everything and
// smaller text, so that a screen of numbers fits a window without scrolling, and the window is small enough to sit
// beside the game. The colours of the pages are picked for a dark background, and are hard to read on a light one.
func Theme() fyne.Theme { return denseTheme{theme.DefaultTheme()} }

type denseTheme struct{ fyne.Theme }

// Color lightens the quiet text of the dark theme, which is too dark to read at the size the numbers are drawn.
func (t denseTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameDisabled {
		return color.NRGBA{R: 0x9a, G: 0x9a, B: 0x9a, A: 0xff}
	}
	return t.Theme.Color(name, theme.VariantDark)
}

func (t denseTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameInnerPadding:
		return 4
	case theme.SizeNamePadding:
		return 3
	case theme.SizeNameText:
		return 12
	}
	return t.Theme.Size(name)
}
