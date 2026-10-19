package playerui_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	player "app/internal/core/player/domain"
	"app/internal/delivery/ui/playerui"
)

var _ = Describe("playerui presenter", func() {
	It("shows the breeder's name, or says that the game does not keep it where it can be read", func() {
		Expect(playerui.BreederText(player.Player{Breeder: "Bob"})).To(Equal("Bob"))
		Expect(playerui.BreederText(player.Player{})).To(Equal("(this game keeps the breeder's name where the viewer cannot read it)"))
	})

	It("groups the buttons held as the shoulders, the face, the middle and the pad", func() {
		// L1, R1 | Circle, Square | Start | Up, Left
		input := 1<<2 | 1<<3 | 1<<4 | 1<<7 | 1<<11 | 1<<12 | 1<<15
		Expect(playerui.NewControllerView(player.Player{Input: input})).To(Equal(playerui.ControllerView{
			Shoulder: "L1 R1", Face: "Circle Square", Misc: "Start", DPad: "Up Left",
		}))
	})

	It("shows no buttons held as nothing", func() {
		Expect(playerui.NewControllerView(player.Player{})).To(Equal(playerui.ControllerView{}))
	})
})
