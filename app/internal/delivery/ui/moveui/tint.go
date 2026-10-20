package moveui

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"app/internal/core/move/domain"
)

// The game draws the icon of a move that runs on power in yellow, and one that runs on intelligence in green. The icons
// here are black, white and grey, so the white of the fill is what is given the colour.
var (
	powerFill        = color.NRGBA{R: 255, G: 222, B: 48, A: 255}
	intelligenceFill = color.NRGBA{R: 64, G: 224, B: 96, A: 255}
)

// tint gives an icon the colour of the kind of its move, in place of its white, and leaves the black and the grey as
// they are. An icon of a move of no known kind is as it was.
func tint(icon []byte, kind domain.Kind) ([]byte, error) {
	var fill color.NRGBA
	switch kind {
	case domain.KindPower:
		fill = powerFill
	case domain.KindIntelligence:
		fill = intelligenceFill
	default:
		return icon, nil
	}
	src, err := png.Decode(bytes.NewReader(icon))
	if err != nil {
		return nil, err
	}
	img := image.NewNRGBA(src.Bounds())
	draw.Draw(img, img.Bounds(), src, src.Bounds().Min, draw.Src)
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			if c := img.NRGBAAt(x, y); c.R == 255 && c.G == 255 && c.B == 255 && c.A == 255 {
				img.SetNRGBA(x, y, fill)
			}
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
