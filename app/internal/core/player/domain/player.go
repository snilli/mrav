package domain

import monster "app/internal/core/monster/domain"

const (
	ItemSlots         = 20
	ButtonCount       = 16
	errantrySaleReady = 255
)

var ButtonNames = [ButtonCount]string{
	"L2", "R2", "L1", "R1", "Circle", "Cross", "Triangle", "Square",
	"Select", "L3", "R3", "Start", "Up", "Right", "Down", "Left",
}

type Player struct {
	Money      int
	NextSale   int
	ErrantryCD int
	Items      [ItemSlots]int
	Breeder    string
	Input      int
	Ranch      Ranch
	Version    monster.Version
}

func (p Player) ErrantrySaleReady() bool { return p.ErrantryCD == errantrySaleReady }

func (p Player) Held() [ButtonCount]bool {
	var held [ButtonCount]bool
	for b := range held {
		held[b] = p.Input&(1<<b) != 0
	}
	return held
}
