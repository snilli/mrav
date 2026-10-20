package moveui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"

	"app/internal/core/move/domain"
)

// The mark of a move that asks for a good nature or a bad one is drawn here, at the size of the game's marks: a light
// tile with brows for a good nature, a dark one with yellow eyes for a bad one.
const (
	natureW = 24
	natureH = 16
)

// A part of a face is the pixels of its left side, as column and row; the right side is the reflection of it.
type pixelSet [][2]int

var (
	brows = pixelSet{ // an arch over each eye
		{6, 6}, {7, 6}, {8, 6}, {5, 7}, {6, 7}, {8, 7}, {9, 7}, {4, 8}, {5, 8}, {9, 8}, {10, 8}, {4, 9}, {10, 9},
	}
	eyes = pixelSet{ // a slit that slopes down toward the middle
		{4, 6}, {5, 6}, {5, 7}, {6, 7}, {6, 8}, {7, 8}, {7, 9}, {8, 9}, {8, 10}, {9, 10}, {9, 11}, {10, 11},
	}
)

// natureIcon is the PNG of the mark for a nature, or nil for a move that asks for none.
func natureIcon(n domain.Nature) []byte {
	var ground, edge, feature color.NRGBA
	var face pixelSet
	switch n {
	case domain.NatureGood:
		ground, edge, feature, face = color.NRGBA{R: 246, G: 246, B: 246, A: 255}, color.NRGBA{R: 140, G: 140, B: 140, A: 255}, color.NRGBA{R: 60, G: 60, B: 60, A: 255}, brows
	case domain.NatureBad:
		ground, edge, feature, face = color.NRGBA{R: 40, G: 40, B: 40, A: 255}, color.NRGBA{R: 110, G: 110, B: 110, A: 255}, color.NRGBA{R: 255, G: 232, B: 48, A: 255}, eyes
	default:
		return nil
	}
	img := image.NewNRGBA(image.Rect(0, 0, natureW, natureH))
	for y := range natureH {
		for x := range natureW {
			// the corners are cut, to make the tile round
			if x+y < 3 || natureW-1-x+y < 3 || x+natureH-1-y < 3 || natureW-1-x+natureH-1-y < 3 {
				continue
			}
			c := ground
			if x == 0 || y == 0 || x == natureW-1 || y == natureH-1 {
				c = edge
			}
			img.SetNRGBA(x, y, c)
		}
	}
	for _, p := range face {
		img.SetNRGBA(p[0], p[1], feature)
		img.SetNRGBA(natureW-1-p[0], p[1], feature)
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		panic(err) // an image in memory that cannot be written is a mistake in the program
	}
	return out.Bytes()
}
