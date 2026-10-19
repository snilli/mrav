// Package domain is the player's side of the game: whatever is not the monster. Pure Go.
package domain

const (
	// ItemSlots is how many items the player can carry.
	ItemSlots = 20
	// ButtonCount is how many buttons a controller reports.
	ButtonCount = 16
	// errantrySaleReady is what the errantry countdown reads when its sale is ready.
	errantrySaleReady = 255
)

// ButtonNames are the buttons of controller 1, in the order of their bits in Player.Input.
var ButtonNames = [ButtonCount]string{
	"L2", "R2", "L1", "R1", "Circle", "Cross", "Triangle", "Square",
	"Select", "L3", "R3", "Start", "Up", "Right", "Down", "Left",
}

// Player is a snapshot of the game's state that is not the monster.
type Player struct {
	Money      int
	NextSale   int // weeks until the shop's next sale
	ErrantryCD int // the errantry sale's countdown; it reads 255 when the sale is ready
	Items      [ItemSlots]int
	Breeder    string // the breeder's name; empty when the game keeps it where this viewer cannot read it
	Input      int    // the buttons held on controller 1, one bit each
	Ranch      Ranch
}

// ErrantrySaleReady says whether the errantry sale can be had now.
func (p Player) ErrantrySaleReady() bool { return p.ErrantryCD == errantrySaleReady }

// Held says which buttons of controller 1 are down, in the order of ButtonNames.
func (p Player) Held() [ButtonCount]bool {
	var held [ButtonCount]bool
	for b := range held {
		held[b] = p.Input&(1<<b) != 0
	}
	return held
}
