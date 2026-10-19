// Package domain is a genus's moves: what each one is, and when a monster may learn it.
package domain

import monster "app/internal/core/monster/domain"

// Kind is whether a move runs on Power or on Intelligence.
type Kind string

const (
	KindPower        Kind = "power"
	KindIntelligence Kind = "intelligence"
	// KindUnknown is a move the source never states a kind for. It is a value, not a guess.
	KindUnknown Kind = "unknown"
)

// Scale picks which of the game's rating tables turns a raw value into words (the first argument of
// the game's GenerateStatValue).
type Scale int

const (
	ScaleForce Scale = iota
	ScaleHit
	ScaleWither
	ScaleCrit
)

// Extra is the optional row of a move such as "HP Drain: 100%". Text is a literal; Scale and Raw name a
// value the game rates in words.
type Extra struct {
	Label string
	Text  string
	Scale *Scale
	Raw   int
}

// Move is one technique in a genus's 4x6 grid.
type Move struct {
	Genus      monster.Genus
	MemorySlot int // index into the game's 24-entry learned and used arrays
	Col, Row   int // 1-based cell in the grid
	Name       string
	Kind       Kind
	Guts       int
	// Force, Hit, Wither and Crit are raw values; 0 reads "---" (Hit is always shown).
	Force, Hit, Wither, Crit int
	Extra                    *Extra
	Description              string
	Unlock                   UnlockRule
	Icon                     string // asset key, "Pixie_1"
}
