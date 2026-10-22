package shellui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func Theme() fyne.Theme { return denseTheme{theme.DefaultTheme()} }

type denseTheme struct{ fyne.Theme }

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
