package gamememory

import monster "app/internal/core/monster/domain"

type Text struct {
	Addr  uint32
	Len   int
	Steam bool
}

type Layout struct {
	Genus, SubGenus uint32
	Age             uint32
	Stats           [monster.StatCount]uint32
	Lifespan        uint32
	InitialLifespan uint32
	Nature          uint32
	NatureBase      uint32
	Fatigue, Fame   uint32
	Stress          uint32
	Spoil, Fear     uint32
	Form            uint32
	Growth          uint32
	TrainBoost      uint32
	CupJellies      uint32
	GoldPeach       uint32
	SilverPeach     uint32
	Motivation      uint32
	PlayType        uint32
	Drug            uint32
	DrugWeeks       uint32
	ItemUsed        uint32
	ItemLike        uint32
	ItemDislike     uint32
	Rank            uint32
	LifeStage       uint32
	LifeType        uint32
	Moves           uint32
	Active          uint32
	ArenaSpeed      uint32
	GutsRate        uint32
	Specials        uint32
	PrizeMoney      uint32
	Name            Text

	Money      uint32
	Items      uint32
	NextSale   uint32
	ErrantryCD uint32
	Breeder    Text
	Input      uint32

	Record                    uint32
	Date, Tournament          uint32
	Grade, House, Barn, Altar uint32
	Unlocks, Phoenix          uint32
	Energy                    uint32
	Tile, Position            uint32
	WalkInRAM                 bool
	PatchedSwim               bool
}

var (
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
		Breeder: Text{Addr: 0x98FAA, Len: 12},
		Record:  0x97A66,
		Date:    0x99228, Tournament: 0x99080, Grade: 0x98FC0, House: 0x98FC1, Barn: 0x98FC2, Altar: 0x99050,
		Unlocks: 0x99386, Phoenix: 0x993CE,
		Energy: 0x0D55BC, Tile: 0x0D55B6, Position: 0x0D55C0, WalkInRAM: true,
	}

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
		Unlocks: 0x9A936, Energy: 0x1DE8B84, Tile: 0x1DE8B7E, Position: 0x1DE8B88,
		PatchedSwim: true,
	}
)
