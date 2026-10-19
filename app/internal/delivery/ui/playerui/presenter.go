// Package playerui is the Breeder and Controller pages: what the game keeps of the player.
package playerui

import (
	"strings"

	player "app/internal/core/player/domain"
)

// BreederText is the breeder's name, or says why there is none to show.
func BreederText(p player.Player) string {
	if p.Breeder == "" {
		return "(this game keeps the breeder's name where the viewer cannot read it)"
	}
	return p.Breeder
}

// ControllerView is the buttons held on controller 1, group by group.
type ControllerView struct {
	Shoulder string
	Face     string
	Misc     string
	DPad     string
}

func NewControllerView(p player.Player) ControllerView {
	held := p.Held()
	names := func(from, to int) string {
		var out []string
		for b := from; b < to; b++ {
			if held[b] {
				out = append(out, player.ButtonNames[b])
			}
		}
		return strings.Join(out, " ")
	}
	return ControllerView{Shoulder: names(0, 4), Face: names(4, 8), Misc: names(8, 12), DPad: names(12, 16)}
}
