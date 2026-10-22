package domain

import "fmt"

type band struct {
	upTo  int
	grade string
}

var (
	wide   = []band{{9, "E"}, {19, "D"}, {29, "C"}, {39, "B"}, {49, "A"}}
	narrow = []band{{9, "E"}, {14, "D"}, {19, "C"}, {24, "B"}, {29, "A"}}
)

func Rate(scale Scale, raw int) string {
	switch scale {
	case ScaleHit:
		return rateHit(raw)
	case ScaleCrit:
		return rateByBands(raw, narrow)
	default:
		return rateByBands(raw, wide)
	}
}

func rateByBands(raw int, bands []band) string {
	if raw == 0 {
		return "---"
	}
	if raw > 0 {
		for _, b := range bands {
			if raw <= b.upTo {
				return fmt.Sprintf("%s (%d)", b.grade, raw)
			}
		}
	}
	return fmt.Sprintf("S (%d)", raw)
}

func rateHit(raw int) string {
	wrapped := raw
	for wrapped > 156 {
		wrapped -= 256
	}
	grade := "S"
	switch {
	case wrapped <= -15:
		grade = "E"
	case wrapped >= -14 && wrapped <= -5:
		grade = "D"
	case wrapped >= -4 && wrapped <= 0:
		grade = "C"
	case raw >= 1 && raw <= 4:
		grade = "B"
	case raw >= 5 && raw <= 14:
		grade = "A"
	}
	return fmt.Sprintf("%s (%d)", grade, raw)
}
