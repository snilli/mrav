package gamememory

import (
	"encoding/binary"
	"math"
	"slices"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
)

const (
	noMonster = "No monster"
	noYear    = 0xFFFF
)

type reader struct {
	mem Memory
	err error
}

func (r *reader) read(addr uint32, n int) []byte {
	p := make([]byte, n)
	if r.err == nil {
		r.err = r.mem.Read(addr, p)
	}
	return p
}

func (r *reader) u8(addr uint32) int  { return int(r.read(addr, 1)[0]) }
func (r *reader) u16(addr uint32) int { return int(binary.LittleEndian.Uint16(r.read(addr, 2))) }
func (r *reader) i16(addr uint32) int { return int(int16(r.u16(addr))) }
func (r *reader) i32(addr uint32) int {
	return int(int32(binary.LittleEndian.Uint32(r.read(addr, 4))))
}

func natureSwing(nature int) int {
	sine := math.Round(math.Sin(math.Pi*float64(nature)/2048) * 4096)
	return int(math.Round(sine * 100 / 4096))
}

func decode(l Layout, mem Memory) (monster.Monster, error) {
	r := &reader{mem: mem}

	genus, sub := r.u8(l.Genus), r.u8(l.SubGenus)
	growth := r.read(l.Growth, monster.StatCount)
	guts := r.u8(l.GutsRate)
	if r.err != nil {
		return monster.Monster{}, r.err
	}
	blank := genus == 0 && sub == 0 || genus == 46 && sub == 46
	if guts == 0 && (blank || slices.Contains(growth, 0)) {
		return monster.Monster{}, apperr.NotFound("monster", noMonster)
	}

	natureBase := int(int8(r.u8(l.NatureBase)))
	swing := natureSwing(r.i16(l.Nature))

	fatigue := r.u8(l.Fatigue)
	switch {
	case fatigue > 100 && fatigue < 128:
		fatigue = 100
	case fatigue > 128:
		fatigue = 0
	}

	m := monster.Monster{
		Genus:           monster.Genus(genus),
		SubGenus:        monster.Genus(sub),
		Nature:          natureBase + swing,
		NatureBase:      natureBase,
		PatchedSwim:     l.PatchedSwim,
		Age:             r.i16(l.Age),
		Lifespan:        r.i16(l.Lifespan),
		InitialLifespan: r.i16(l.InitialLifespan),
		Fatigue:         fatigue,
		Fame:            r.u8(l.Fame),
		Stress:          r.u8(l.Stress),
		Form:            int(int8(r.u8(l.Form))),
		Spoil:           r.u8(l.Spoil),
		Fear:            r.u8(l.Fear),
		TrainBoost:      r.u16(l.TrainBoost),
		CupJellies:      r.u8(l.CupJellies),
		GoldPeach:       r.u8(l.GoldPeach) != 0,
		SilverPeach:     r.u8(l.SilverPeach) != 0,
		ItemUsed:        r.u8(l.ItemUsed) != 0,
		PlayType:        r.u8(l.PlayType),
		Drug:            r.u8(l.Drug),
		DrugWeeks:       r.u8(l.DrugWeeks),
		ItemLike:        r.u8(l.ItemLike),
		ItemDislike:     r.u8(l.ItemDislike),
		Rank:            r.u8(l.Rank),
		LifeStage:       r.u8(l.LifeStage),
		LifeType:        r.u8(l.LifeType),
		ArenaSpeed:      r.u8(l.ArenaSpeed),
		GutsRate:        guts,
		Specials:        r.u16(l.Specials),
		PrizeMoney:      r.i32(l.PrizeMoney),
		Name:            readText(r, l.Name),
	}
	for s := range m.Stats {
		m.Stats[s] = r.i16(l.Stats[s])
	}
	rec := r.read(l.Record, 16)
	count := func(i int) int { return int(binary.LittleEndian.Uint16(rec[2*i:])) }
	m.Record = monster.Record{
		Wins: count(0), Losses: count(1), KOs: count(2),
		Injuries: count(4), Runaways: count(5), Sicknesses: count(6), Resistance: count(7),
	}
	m.Growth[monster.Pow], m.Growth[monster.Int], m.Growth[monster.Life] = int(growth[0]), int(growth[1]), int(growth[2])
	m.Growth[monster.Skill], m.Growth[monster.Speed], m.Growth[monster.Defense] = int(growth[3]), int(growth[4]), int(growth[5])
	for d, percent := range r.read(l.Motivation, monster.DrillCount) {
		m.Motivation[d] = int(percent)
	}
	moves := r.read(l.Moves, 2*monster.SlotCount)
	for slot := range m.Learned {
		m.Learned[slot] = moves[2*slot] == 1
		m.Used[slot] = int(moves[2*slot+1])
	}
	for g, slot := range r.read(l.Active, monster.ActiveGroups) {
		m.Active[g] = int(slot)
		if m.Active[g] >= monster.SlotCount {
			m.Active[g] = monster.NoActiveMove
		}
	}
	if r.err != nil {
		return monster.Monster{}, r.err
	}
	return m, nil
}

func decodePlayer(l Layout, mem Memory) (player.Player, error) {
	r := &reader{mem: mem}
	p := player.Player{
		Money:      r.i32(l.Money),
		NextSale:   r.u8(l.NextSale),
		ErrantryCD: r.u8(l.ErrantryCD),
		Breeder:    readText(r, l.Breeder),
		Input:      r.u16(l.Input),
	}
	for slot := range p.Items {
		p.Items[slot] = r.u8(l.Items + 4*uint32(slot))
	}
	if l.Date != 0 {
		tournament := r.u16(l.Tournament)
		if tournament == noYear {
			tournament = 0
		}
		p.Ranch = player.Ranch{
			Known: true, Date: r.i32(l.Date), Tournament: tournament, Grade: r.u8(l.Grade),
			House: r.u8(l.House), Barn: r.u8(l.Barn), Altar: r.u8(l.Altar),
		}
		for breed, unlocked := range r.read(l.Unlocks, player.BreedCount) {
			p.Ranch.Unlocked[breed] = unlocked != 0
		}
		if l.Phoenix != 0 {
			p.Ranch.PhoenixKnown, p.Ranch.PhoenixDone = true, r.u16(l.Phoenix) != 0
		}
		var read func(addr uint32, p []byte) error
		if exe, ok := mem.(ModuleMemory); ok && !l.WalkInRAM {
			read = exe.ReadModule
		} else if l.WalkInRAM {
			read = mem.Read
		}
		if read != nil {
			word := func(addr uint32) (uint16, bool) {
				var b [2]byte
				if addr == 0 || read(addr, b[:]) != nil {
					return 0, false
				}
				return binary.LittleEndian.Uint16(b[:]), true
			}
			if energy, ok := word(l.Energy); ok {
				p.Ranch.Energy, p.Ranch.EnergyKnown = int(int16(energy)), true
			}
			if tile, ok := word(l.Tile); ok {
				p.Ranch.Tile, p.Ranch.WalkKnown = int(tile), true
			}
			if x, ok := word(l.Position); ok {
				if y, ok := word(l.Position + 2); ok {
					p.Ranch.OffsetX, p.Ranch.OffsetY, p.Ranch.OffsetKnown = int(int16(x)), int(int16(y)), true
				}
			}
		}
	}
	if r.err != nil {
		return player.Player{}, r.err
	}
	return p, nil
}
