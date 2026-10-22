package moveui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"

	"app/internal/core/move/domain"
)

const (
	natureW = 24
	natureH = 16
)

type pixelSet [][2]int

var (
	brows = pixelSet{
		{6, 6}, {7, 6}, {8, 6}, {5, 7}, {6, 7}, {8, 7}, {9, 7}, {4, 8}, {5, 8}, {9, 8}, {10, 8}, {4, 9}, {10, 9},
	}
	eyes = pixelSet{
		{4, 6}, {5, 6}, {5, 7}, {6, 7}, {6, 8}, {7, 8}, {7, 9}, {8, 9}, {8, 10}, {9, 10}, {9, 11}, {10, 11},
	}
)

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
		panic(err)
	}
	return out.Bytes()
}
