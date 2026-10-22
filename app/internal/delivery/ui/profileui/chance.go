package profileui

import (
	"math"
	"strconv"
)

func ChanceText(percent float64) string {
	return strconv.FormatFloat(math.Round(percent*10)/10, 'f', -1, 64) + "%"
}
