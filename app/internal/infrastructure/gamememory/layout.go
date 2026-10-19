// Package gamememory is a MonsterSource and a PlayerSource that read a running game: MR2 under DuckStation,
// or Monster Rancher 1&2 DX on Steam. Every address in it comes from MR2AV, the viewer this one is ported from.
package gamememory

import monster "app/internal/core/monster/domain"

// Text is a name the game keeps in its RAM: where it starts, how many bytes it can take, and which of the
// game's two encodings it is in. A Text with no address is a name the game does not keep.
type Text struct {
	Addr  uint32
	Len   int
	Steam bool
}

// Layout says where a game keeps what the viewer shows. Every address is an offset into the PlayStation's RAM, as
// MR2AV gives them. Fields are unsigned bytes unless said otherwise.
type Layout struct {
	Genus, SubGenus uint32
	Age             uint32                    // two bytes, signed
	Stats           [monster.StatCount]uint32 // two bytes each, signed
	Lifespan        uint32                    // two bytes, signed
	InitialLifespan uint32                    // two bytes, signed
	Nature          uint32                    // two bytes, signed
	NatureBase      uint32                    // signed
	Fatigue, Fame   uint32
	Stress          uint32
	Spoil, Fear     uint32
	Form            uint32 // signed
	Growth          uint32 // six bytes: Power, Intelligence, Life, Skill, Speed, Defence
	TrainBoost      uint32 // two bytes, flags
	CupJellies      uint32
	GoldPeach       uint32
	SilverPeach     uint32
	Motivation      uint32 // ten bytes: Domino, Study, Run, Shoot, Dodge, Endure, Pull, Meditate, Leap, Swim
	PlayType        uint32
	Drug            uint32
	DrugWeeks       uint32
	ItemUsed        uint32
	ItemLike        uint32
	ItemDislike     uint32
	Rank            uint32
	LifeStage       uint32
	LifeType        uint32
	Moves           uint32 // 24 pairs of bytes: 1 when the move is learned, then the times it was used
	Active          uint32 // four bytes: the memory slot of each column's active move
	ArenaSpeed      uint32
	GutsRate        uint32 // with nothing in it no monster is loaded
	Specials        uint32 // two bytes, flags
	PrizeMoney      uint32 // four bytes, signed
	Name            Text

	Money      uint32 // four bytes, signed
	Items      uint32 // 20 item ids, one byte each, four bytes apart
	NextSale   uint32
	ErrantryCD uint32
	Breeder    Text
	Input      uint32 // two bytes, flags: the buttons held on controller 1

	// What a Cheat Engine table for MR2DX adds to MR2AV's. Record is where the monster's wins are, 16 bytes: wins,
	// losses and knock-outs, two bytes each, a gap of two, then injuries, runaways, sicknesses and resistance. The
	// ranch's addresses are 0 where the game's map is not known, which is the whole of the PlayStation's.
	Record                    uint32
	Date, Tournament          uint32 // four bytes each: the weeks since the game began, and the year of the first IMA-FIMBA tournament
	Grade, House, Barn, Altar uint32 // a byte each
	Unlocks                   uint32 // a byte for each main breed, by its code: 0 while it is locked
	Energy                    uint32 // two bytes, signed, counted from the exe and not the RAM: an expedition's energy
	Tile, Steps               uint32 // two bytes each, counted from the exe too: the tile the monster is on, and its steps
}

var (
	// ps1 is MR2 as the PlayStation version keeps it, vanilla and hard mode, with its English names.
	ps1 = Layout{
		Genus: 0x97A18, SubGenus: 0x97A1C, Age: 0x97A14,
		Stats: [monster.StatCount]uint32{
			monster.Life: 0x97A20, monster.Pow: 0x97A22, monster.Int: 0x97A2A,
			monster.Skill: 0x97A26, monster.Speed: 0x97A28, monster.Defense: 0x97A24,
		},
		Lifespan: 0x97A30, InitialLifespan: 0x97A32, Nature: 0x97A34, NatureBase: 0x97A59,
		Fatigue: 0x97A37, Fame: 0x97A38, Stress: 0x97A3B, Spoil: 0x97A3C, Fear: 0x97A3D, Form: 0x97A3E,
		Growth: 0x97A40, TrainBoost: 0x97A78, CupJellies: 0x97AC5, GoldPeach: 0x97AC6, SilverPeach: 0x97AC7,
		Motivation: 0x97AF0, PlayType: 0x97AFD,
		Drug: 0x97B0A, DrugWeeks: 0x97B0C, ItemUsed: 0x97B0E, ItemLike: 0x97B1A, ItemDislike: 0x97B1C,
		Rank: 0x97B92, LifeStage: 0x97B93, LifeType: 0x97B94,
		Moves: 0x97B9A, Active: 0x97BD0, ArenaSpeed: 0x97BD6, GutsRate: 0x97BD7, Specials: 0x97BD8, PrizeMoney: 0x97A5C,
		Name: Text{Addr: 0x97B78, Len: 12},

		Money: 0x98FBC, Items: 0x9923C, NextSale: 0x992DE, ErrantryCD: 0x992E9, Input: 0x938AE,
		Record: 0x97A66, // the Steam address and 8, as every field of the monster before its name is
	}

	// steam is MR2 inside the English release of Monster Rancher 1&2 DX.
	steam = Layout{
		Genus: 0x97A10, SubGenus: 0x97A14, Age: 0x97A0C,
		Stats: [monster.StatCount]uint32{
			monster.Life: 0x97A18, monster.Pow: 0x97A1A, monster.Int: 0x97A22,
			monster.Skill: 0x97A1E, monster.Speed: 0x97A20, monster.Defense: 0x97A1C,
		},
		Lifespan: 0x97A28, InitialLifespan: 0x97A2A, Nature: 0x97A2C, NatureBase: 0x97A51,
		Fatigue: 0x97A2F, Fame: 0x97A30, Stress: 0x97A33, Spoil: 0x97A34, Fear: 0x97A35, Form: 0x97A36,
		Growth: 0x97A38, TrainBoost: 0x97A70, CupJellies: 0x97ABD, GoldPeach: 0x97ABE, SilverPeach: 0x97ABF,
		Motivation: 0x97AE8, PlayType: 0x97AF5,
		Drug: 0x97B02, DrugWeeks: 0x97B04, ItemUsed: 0x97B06, ItemLike: 0x97B12, ItemDislike: 0x97B14,
		Rank: 0x97B9A, LifeStage: 0x97B9B, LifeType: 0x97B9C,
		Moves: 0x97BA2, Active: 0x97BD8, ArenaSpeed: 0x97BDE, GutsRate: 0x97BDF, Specials: 0x97BE0, PrizeMoney: 0x97A54,
		Name: Text{Addr: 0x97B78, Len: 24, Steam: true},

		Money: 0x9A568, Items: 0x9A7EC, NextSale: 0x9A88E, ErrantryCD: 0x9A899, Input: 0x938AE,
		Breeder: Text{Addr: 0x9A54A, Len: 24, Steam: true},
		Record:  0x97A5E,
		Date:    0x9A7D8, Tournament: 0x9A62C, Grade: 0x9A56C, House: 0x9A56D, Barn: 0x9A56E, Altar: 0x9A5FC,
		Unlocks: 0x9A936, Energy: 0x1DE8B84, Tile: 0x1DE8B7E, Steps: 0x1DE8B8A,
	}
)
