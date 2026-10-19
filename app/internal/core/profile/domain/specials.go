package domain

import "strings"

// Specials are the battle specials a monster has, as the names to show.
type Specials struct {
	Names   []string
	Invalid bool // the monster has a special its breed cannot have
}

// Text is the names, one after the other.
func (s Specials) Text() string { return strings.Join(s.Names, ", ") }

const (
	specialPower = 1
	specialAnger = 2
	specialReal  = 1024
	specialDrunk = 2048
	specialUnity = 4096
	unusedFlags  = 8192 | 16384 | 32768
)

// specials are the flags in the order the game lists them.
var specials = [...]struct {
	flag int
	name string
}{
	{specialPower, "Power"}, {specialAnger, "Anger"}, {4, "Grit"}, {8, "Will"}, {16, "Fight"}, {32, "Fury"},
	{64, "Guard"}, {128, "Ease"}, {256, "Hurry"}, {512, "Vigor"},
	{specialReal, "Real"}, {specialDrunk, "Drunk"}, {specialUnity, "Unity"},
}

// SpecialsOf lists the specials in the flags. Power shows in brackets when the effective nature is 20 or more,
// and Anger when it is -20 or less, because that is when they work. Real is for Joker, Drunk for Kato and Unity
// for Pandora; on any other breed, and for the flags the game does not use, the monster is marked invalid.
func SpecialsOf(flags, nature, main, sub int) Specials {
	var s Specials
	for _, sp := range specials {
		if flags&sp.flag == 0 {
			continue
		}
		name := sp.name
		switch sp.flag {
		case specialPower:
			if nature >= 20 {
				name = "[" + name + "]"
			}
		case specialAnger:
			if nature <= -20 {
				name = "[" + name + "]"
			}
		case specialReal:
			s.Invalid = s.Invalid || main != 26 && sub != 26
		case specialDrunk:
			s.Invalid = s.Invalid || main != 16
		case specialUnity:
			s.Invalid = s.Invalid || main != 3
		}
		s.Names = append(s.Names, name)
	}
	if flags&unusedFlags != 0 {
		s.Invalid = true
		s.Names = append(s.Names, "Unused Special")
	}
	return s
}
