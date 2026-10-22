package domain

import monster "app/internal/core/monster/domain"

type Gate struct {
	Stats    []monster.Stat
	Min, Max int
}

type Chain struct{ Slot, Uses int }

type Acquire struct {
	Place, Priority int
	Auto            int
	Gates           []Gate
	Nature          int
	Chain           *Chain
	SubIn, SubNotIn []monster.Genus
}
