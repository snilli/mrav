package domain

import monster "app/internal/core/monster/domain"

type Kind string

const (
	KindPower        Kind = "power"
	KindIntelligence Kind = "intelligence"
	KindUnknown      Kind = "unknown"
)

type Scale int

const (
	ScaleForce Scale = iota
	ScaleHit
	ScaleWither
	ScaleCrit
)

type Extra struct {
	Label string
	Text  string
	Raw   int
}

type Move struct {
	Genus                    monster.Genus
	MemorySlot               int
	Col, Row                 int
	Name                     string
	Kind                     Kind
	Guts                     int
	Force, Hit, Wither, Crit int
	Extra                    *Extra
	Description              string
	Unlock                   UnlockRule
	Acquire                  []Acquire
	Icon                     string
}
