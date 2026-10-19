package gamememory

import (
	"encoding/binary"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"app/internal/core/apperr"
	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
)

// addresses is where MR2AV looks for each field, typed out here from its source (Timer1_Tick, CollateMonMoves,
// ListItems, MonReadMoney and the rest) and not from the layouts, so that a slip in a layout shows as a failing test.
type addresses struct {
	age, genus, sub                                                    uint32
	life, pow, def, skill, speed, intel                                uint32
	lifespan, initial, nature, natureBase                              uint32
	fatigue, fame, stress, spoil, fear, form                           uint32
	growth, trainBoost, cupJellies, gold, silver, motivation, playType uint32
	drug, drugWeeks, itemUsed, like, dislike                           uint32
	rank, lifeStage, lifeType                                          uint32
	moves, active, arena, guts, specials, prize                        uint32
	name, breeder                                                      uint32
	money, items, nextSale, errantry, input                            uint32
	// From a Cheat Engine table for MR2DX, which MR2AV does not read. record is where wins are, with losses and
	// knock-outs after it, then a gap, then injuries, runaways, sicknesses and resistance, two bytes each. The
	// ranch's are known for Steam only.
	record                                      uint32
	date, tournament, grade, house, barn, altar uint32
	unlocks                                     uint32 // a byte for each main breed, by its code: 1 when it is unlocked
	energy, tile, steps                         uint32 // where the expedition's energy, tile and steps are, counted from the exe, not the RAM
}

var (
	// MR2 on the PlayStation, vanilla and hard mode.
	ps1MR2AV = addresses{
		age: 0x97A14, genus: 0x97A18, sub: 0x97A1C,
		life: 0x97A20, pow: 0x97A22, def: 0x97A24, skill: 0x97A26, speed: 0x97A28, intel: 0x97A2A,
		lifespan: 0x97A30, initial: 0x97A32, nature: 0x97A34, natureBase: 0x97A59,
		fatigue: 0x97A37, fame: 0x97A38, stress: 0x97A3B, spoil: 0x97A3C, fear: 0x97A3D, form: 0x97A3E,
		growth: 0x97A40, trainBoost: 0x97A78, cupJellies: 0x97AC5, gold: 0x97AC6, silver: 0x97AC7,
		motivation: 0x97AF0, playType: 0x97AFD,
		drug: 0x97B0A, drugWeeks: 0x97B0C, itemUsed: 0x97B0E, like: 0x97B1A, dislike: 0x97B1C,
		rank: 0x97B92, lifeStage: 0x97B93, lifeType: 0x97B94,
		moves: 0x97B9A, active: 0x97BD0, arena: 0x97BD6, guts: 0x97BD7, specials: 0x97BD8, prize: 0x97A5C,
		name: 0x97B78, breeder: 0,
		money: 0x98FBC, items: 0x9923C, nextSale: 0x992DE, errantry: 0x992E9, input: 0x938AE,
		record: 0x97A66, // the Steam one and 8, as every field of the monster before its name is
	}
	// MR2DX on Steam, English release.
	steamMR2AV = addresses{
		age: 0x97A0C, genus: 0x97A10, sub: 0x97A14,
		life: 0x97A18, pow: 0x97A1A, def: 0x97A1C, skill: 0x97A1E, speed: 0x97A20, intel: 0x97A22,
		lifespan: 0x97A28, initial: 0x97A2A, nature: 0x97A2C, natureBase: 0x97A51,
		fatigue: 0x97A2F, fame: 0x97A30, stress: 0x97A33, spoil: 0x97A34, fear: 0x97A35, form: 0x97A36,
		growth: 0x97A38, trainBoost: 0x97A70, cupJellies: 0x97ABD, gold: 0x97ABE, silver: 0x97ABF,
		motivation: 0x97AE8, playType: 0x97AF5,
		drug: 0x97B02, drugWeeks: 0x97B04, itemUsed: 0x97B06, like: 0x97B12, dislike: 0x97B14,
		rank: 0x97B9A, lifeStage: 0x97B9B, lifeType: 0x97B9C,
		moves: 0x97BA2, active: 0x97BD8, arena: 0x97BDE, guts: 0x97BDF, specials: 0x97BE0, prize: 0x97A54,
		name: 0x97B78, breeder: 0x9A54A,
		money: 0x9A568, items: 0x9A7EC, nextSale: 0x9A88E, errantry: 0x9A899, input: 0x938AE,
		record: 0x97A5E,
		date:   0x9A7D8, tournament: 0x9A62C, grade: 0x9A56C, house: 0x9A56D, barn: 0x9A56E, altar: 0x9A5FC,
		unlocks: 0x9A936,
		energy:  0x1DE8B84, tile: 0x1DE8B7E, steps: 0x1DE8B8A,
	}
)

const ramSize = 0x200000 // the PlayStation's 2 MB

// image is a game's RAM, which can be made to fail, and which counts how often it is let go of. It also has the little
// that a game keeps in its exe, outside the RAM, which can be made to fail by itself.
type image struct {
	ram        []byte
	fail       error
	closed     int
	module     map[uint32]byte
	moduleFail error
}

// ReadModule reads from the exe: what was put there with i16module, and nothing for the rest.
func (i *image) ReadModule(offset uint32, p []byte) error {
	if i.moduleFail != nil {
		return i.moduleFail
	}
	for n := range p {
		p[n] = i.module[offset+uint32(n)]
	}
	return nil
}

func (i *image) i16module(offset uint32, v int16) {
	if i.module == nil {
		i.module = map[uint32]byte{}
	}
	i.module[offset], i.module[offset+1] = byte(uint16(v)), byte(uint16(v)>>8)
}

func newImage() *image { return &image{ram: make([]byte, ramSize)} }

func (i *image) Read(addr uint32, p []byte) error {
	if i.fail != nil {
		return i.fail
	}
	copy(p, i.ram[addr:])
	return nil
}

func (i *image) Close() error {
	i.closed++
	return nil
}

func (i *image) u8(addr uint32, v byte)     { i.ram[addr] = v }
func (i *image) put(addr uint32, b ...byte) { copy(i.ram[addr:], b) }
func (i *image) i16(addr uint32, v int16) {
	binary.LittleEndian.PutUint16(i.ram[addr:], uint16(v))
}
func (i *image) i32(addr uint32, v int32) {
	binary.LittleEndian.PutUint32(i.ram[addr:], uint32(v))
}

// The names the images spell, in each game's own codes: "Aa 1", and for the Steam breeder "Bob".
var (
	ps1Name     = []byte{11, 37, 0, 2, 255}
	steamName   = []byte{0xB0, 0x00, 0xB0, 0x1A, 0x00, 0xA3, 0xFF}
	steamBreeds = []byte{0xB0, 0x01, 0xB0, 0x28, 0xB0, 0x1B, 0xFF}
)

// monsterImage is a Genus 12 / Sub 3 monster whose every field has a value of its own, in the form the game
// keeps it, and a player with a value in every field too.
func monsterImage(a addresses, name, breeder []byte) *image {
	img := newImage()
	img.u8(a.genus, 12)
	img.u8(a.sub, 3)
	img.i16(a.age, 100)
	img.i16(a.life, 111)
	img.i16(a.pow, 222)
	img.i16(a.def, 333)
	img.i16(a.skill, 444)
	img.i16(a.speed, 555)
	img.i16(a.intel, 666)
	img.i16(a.lifespan, 250)
	img.i16(a.initial, 300)
	img.i16(a.nature, 1024) // sin(pi/2) * 100 = 100
	img.u8(a.natureBase, 30)
	img.u8(a.fatigue, 40)
	img.u8(a.fame, 50)
	img.u8(a.stress, 60)
	img.u8(a.spoil, 70)
	img.u8(a.fear, 80)
	img.u8(a.form, 0xF6)                // -10
	img.put(a.growth, 1, 2, 3, 4, 0, 2) // Power, Intelligence, Life, Skill, Speed, Defence
	img.i16(a.trainBoost, 0x0405)
	img.u8(a.cupJellies, 31)
	img.u8(a.gold, 1)
	img.u8(a.silver, 0)
	img.put(a.motivation, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100)
	img.u8(a.playType, 2)
	img.u8(a.drug, 38)
	img.u8(a.drugWeeks, 3)
	img.u8(a.itemUsed, 1)
	img.u8(a.like, 4)
	img.u8(a.dislike, 130)
	img.u8(a.rank, 2)
	img.u8(a.lifeStage, 5)
	img.u8(a.lifeType, 3)
	img.u8(a.arena, 4)
	img.u8(a.guts, 5)
	img.i16(a.specials, 0x0403)
	img.i32(a.prize, 12345)
	for slot, flagAndUses := range map[uint32][2]byte{3: {1, 77}, 5: {0, 12}, 7: {2, 9}, 23: {1, 255}} {
		img.put(a.moves+2*slot, flagAndUses[0], flagAndUses[1])
	}
	img.put(a.active, 0, 6, 12, 255)
	img.put(a.name, name...)
	img.i16(a.record, 12)   // wins
	img.i16(a.record+2, 3)  // losses
	img.i16(a.record+4, 5)  // knock-outs
	img.i16(a.record+6, -1) // not a count, and not read
	img.i16(a.record+8, 4)  // injuries
	img.i16(a.record+10, 6) // runaways
	img.i16(a.record+12, 7) // sicknesses
	img.i16(a.record+14, 8) // resistance

	img.i32(a.money, 987654)
	for slot := range uint32(player.ItemSlots) {
		img.u8(a.items+4*slot, byte(slot*9))
	}
	img.u8(a.nextSale, 7)
	img.u8(a.errantry, 255)
	if a.breeder != 0 {
		img.put(a.breeder, breeder...)
	}
	if a.date != 0 {
		img.i32(a.date, 155)
		img.i32(a.tournament, 1003)
		img.u8(a.grade, 6)
		img.u8(a.house, 2)
		img.u8(a.barn, 1)
		img.u8(a.altar, 3)
		// Pixie and Dragon unlocked, Henger not, Worm (36) unlocked with any value but 0, Naga (37) not.
		img.put(a.unlocks, 1, 1, 0, 0, 0, 0)
		img.u8(a.unlocks+36, 2)
		img.i16module(a.energy, -5) // below nothing, as the energy of an expedition can go
		img.i16module(a.tile, 911)  // column 15 of row 28, on a map 32 tiles wide
		img.i16module(a.steps, 2)
	}
	img.i16(a.input, 48) // Circle and Cross
	return img
}

// ramOnly is a Memory that is the RAM and nothing else: it cannot read the exe, as the PlayStation's cannot.
type ramOnly struct{ image *image }

func (r *ramOnly) Read(addr uint32, p []byte) error { return r.image.Read(addr, p) }
func (r *ramOnly) Close() error                     { return r.image.Close() }

func reason(err error) string {
	GinkgoHelper()
	var e *apperr.Error
	Expect(errors.As(err, &e)).To(BeTrue(), "%v is not an apperr.Error", err)
	return e.Msg
}

var _ = Describe("decode", func() {
	DescribeTable("reads the monster from where its game keeps it",
		func(l Layout, a addresses, name []byte, wantName string) {
			got, err := decode(l, monsterImage(a, name, steamBreeds))
			Expect(err).NotTo(HaveOccurred())

			want := monster.Monster{
				Genus: 12, SubGenus: 3,
				Stats:  [monster.StatCount]int{111, 222, 666, 444, 555, 333}, // Life, Pow, Int, Skill, Speed, Defense
				Nature: 130, NatureBase: 30,
				Active: [monster.ActiveGroups]int{0, 6, 12, monster.NoActiveMove},
				Name:   wantName,
				Age:    100, Lifespan: 250, InitialLifespan: 300,
				Fatigue: 40, Fame: 50, Stress: 60, Form: -10, Spoil: 70, Fear: 80,
				Growth:     [monster.StatCount]int{3, 1, 2, 4, 0, 2}, // by stat, not in the order the game keeps them
				TrainBoost: 0x0405,
				CupJellies: 31, GoldPeach: true, ItemUsed: true,
				Motivation: [monster.DrillCount]int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100},
				PlayType:   2, Drug: 38, DrugWeeks: 3, ItemLike: 4, ItemDislike: 130,
				Rank: 2, LifeStage: 5, LifeType: 3, ArenaSpeed: 4, GutsRate: 5,
				Specials: 0x0403, PrizeMoney: 12345,
				Record: monster.Record{Wins: 12, Losses: 3, KOs: 5, Injuries: 4, Runaways: 6, Sicknesses: 7, Resistance: 8},
			}
			want.Learned[3], want.Learned[23] = true, true // only a flag of exactly 1 is learned
			want.Used[3], want.Used[5], want.Used[7], want.Used[23] = 77, 12, 9, 255
			Expect(got).To(Equal(want))
		},
		Entry("MR2 on the PlayStation", ps1, ps1MR2AV, ps1Name, "Aa 1"),
		Entry("MR2DX on Steam", steam, steamMR2AV, steamName, "Aa 1"),
	)

	DescribeTable("reads the player's side from where its game keeps it",
		func(l Layout, a addresses, wantBreeder string, wantRanch player.Ranch) {
			got, err := decodePlayer(l, monsterImage(a, ps1Name, steamBreeds))
			Expect(err).NotTo(HaveOccurred())

			want := player.Player{Money: 987654, NextSale: 7, ErrantryCD: 255, Breeder: wantBreeder, Input: 48, Ranch: wantRanch}
			for slot := range want.Items {
				want.Items[slot] = slot * 9
			}
			Expect(got).To(Equal(want))
		},
		Entry("MR2 on the PlayStation, which has no breeder name or ranch to read", ps1, ps1MR2AV, "", player.Ranch{}),
		Entry("MR2DX on Steam", steam, steamMR2AV, "Bob",
			player.Ranch{
				Known: true, Date: 155, Tournament: 1003, Grade: 6, House: 2, Barn: 1, Altar: 3,
				Unlocked: [player.BreedCount]bool{0: true, 1: true, 36: true}, Energy: -5, EnergyKnown: true,
				Tile: 911, Steps: 2, WalkKnown: true,
			}),
	)

	It("still reads the ranch when the exe cannot be read, and says the energy is not known", func() {
		img := monsterImage(steamMR2AV, steamName, steamBreeds)
		img.moduleFail = errors.New("access denied")

		got, err := decodePlayer(steam, img)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Ranch.Known).To(BeTrue())
		Expect(got.Ranch.House).To(Equal(2))
		Expect(got.Ranch.EnergyKnown).To(BeFalse())
		Expect(got.Ranch.Energy).To(BeZero())
		Expect(got.Ranch.WalkKnown).To(BeFalse())
	})

	It("does not look for the energy in a memory that has no exe to read", func() {
		mem := &ramOnly{image: monsterImage(steamMR2AV, steamName, steamBreeds)}
		got, err := decodePlayer(steam, mem)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Ranch.Known).To(BeTrue())
		Expect(got.Ranch.EnergyKnown).To(BeFalse())
		Expect(got.Ranch.WalkKnown).To(BeFalse())
	})

	DescribeTable("works out the nature the way the game shows it: the base plus the sine, cut toward zero",
		func(nature int16, base byte, want int) {
			img := monsterImage(ps1MR2AV, ps1Name, nil)
			img.i16(ps1MR2AV.nature, nature)
			img.u8(ps1MR2AV.natureBase, base)
			got, err := decode(ps1, img)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Nature).To(Equal(want))
			Expect(got.NatureBase).To(Equal(int(int8(base))))
		},
		Entry("nothing at all", int16(0), byte(0), 0),
		Entry("the sine at its top", int16(1024), byte(30), 130),
		Entry("the sine at its bottom, on a negative base", int16(-1024), byte(0xE2), -130),
		Entry("a fraction is dropped: 70.7 is 70", int16(512), byte(0), 70),
		Entry("and toward zero for a negative one: -70.7 is -70", int16(-512), byte(0), -70),
		Entry("a base of 255 is -1", int16(0), byte(0xFF), -1),
	)

	DescribeTable("fixes up the bytes MR2AV fixes up",
		func(fatigue, form byte, wantFatigue, wantForm int) {
			img := monsterImage(ps1MR2AV, ps1Name, nil)
			img.u8(ps1MR2AV.fatigue, fatigue)
			img.u8(ps1MR2AV.form, form)
			got, err := decode(ps1, img)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Fatigue).To(Equal(wantFatigue))
			Expect(got.Form).To(Equal(wantForm))
		},
		Entry("in range", byte(100), byte(100), 100, 100),
		Entry("fatigue over 100 and under 128 reads as 100", byte(101), byte(0), 100, 0),
		Entry("up to 127", byte(127), byte(127), 100, 127),
		Entry("128 is left alone, as MR2AV leaves it", byte(128), byte(128), 128, -128),
		Entry("over 128 reads as no fatigue; form 255 is -1", byte(129), byte(255), 0, -1),
	)

	DescribeTable("says there is no monster when the memory holds none",
		func(main, sub, guts byte, zeroGrowth, none bool) {
			a := ps1MR2AV
			img := monsterImage(a, ps1Name, nil)
			img.put(a.growth, 1, 1, 1, 1, 1, 1) // real growth rates, unless the case says one is nothing
			img.u8(a.genus, main)
			img.u8(a.sub, sub)
			img.u8(a.guts, guts)
			if zeroGrowth {
				img.u8(a.growth+4, 0)
			}
			got, err := decode(ps1, img)
			if !none {
				Expect(err).NotTo(HaveOccurred())
				Expect(got.Genus).To(Equal(monster.Genus(main)))
				return
			}
			Expect(errors.Is(err, apperr.ErrNotFound)).To(BeTrue())
			Expect(reason(err)).To(Equal("No monster"))
		},
		Entry("Pixie and Pixie with no guts rate: nothing is loaded", byte(0), byte(0), byte(0), false, true),
		Entry("46 and 46 with no guts rate: the game's blank", byte(46), byte(46), byte(0), false, true),
		Entry("no guts rate and a growth rate of nothing", byte(12), byte(3), byte(0), true, true),
		Entry("Pixie and Pixie with a guts rate is a Pixie", byte(0), byte(0), byte(5), false, false),
		Entry("no guts rate but real growth rates is still a monster", byte(12), byte(3), byte(0), false, false),
	)

	It("fails with what the memory said when it cannot be read, which is not the same as no monster", func() {
		boom := errors.New("access denied")
		img := monsterImage(ps1MR2AV, ps1Name, nil)
		img.fail = boom

		_, err := decode(ps1, img)
		Expect(errors.Is(err, boom)).To(BeTrue())
		Expect(errors.Is(err, apperr.ErrNotFound)).To(BeFalse())

		_, err = decodePlayer(ps1, img)
		Expect(errors.Is(err, boom)).To(BeTrue())
	})
})
