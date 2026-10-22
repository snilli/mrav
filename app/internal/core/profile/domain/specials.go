package domain

import "strings"

type Specials struct {
	Names   []string
	Invalid bool
}

func (s Specials) Text() string { return strings.Join(s.Names, ", ") }

const (
	specialPower = 1
	specialAnger = 2
	specialReal  = 1024
	specialDrunk = 2048
	specialUnity = 4096
	unusedFlags  = 8192 | 16384 | 32768
)

var specials = [...]struct {
	flag int
	name string
}{
	{specialPower, "Power"}, {specialAnger, "Anger"}, {4, "Grit"}, {8, "Will"}, {16, "Fight"}, {32, "Fury"},
	{64, "Guard"}, {128, "Ease"}, {256, "Hurry"}, {512, "Vigor"},
	{specialReal, "Real"}, {specialDrunk, "Drunk"}, {specialUnity, "Unity"},
}

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
