package demoui

import (
	"math/rand/v2"
	"strconv"

	item "app/internal/core/item/domain"
	monster "app/internal/core/monster/domain"
	"app/internal/core/move/domain"
	player "app/internal/core/player/domain"
)

func randomize(rng *rand.Rand, genera []domain.Genus, m *monster.Monster, p *player.Player) {
	flag := func() bool { return rng.IntN(2) == 0 }

	if len(genera) > 0 {
		m.Genus = genera[rng.IntN(len(genera))].Code
	}
	m.SubGenus = m.Genus
	if rng.IntN(10) < 3 {
		m.SubGenus = monster.Genus(rng.IntN(38))
	}
	m.Name = "Demo " + strconv.Itoa(rng.IntN(100))
	for s := range m.Stats {
		m.Stats[s] = 1 + rng.IntN(999)
		m.Growth[s] = rng.IntN(5)
	}
	m.NatureBase = rng.IntN(101) - 50
	m.Nature = m.NatureBase + rng.IntN(101) - 50
	m.Age = rng.IntN(400)
	m.InitialLifespan = 150 + rng.IntN(250)
	m.Lifespan = 1 + rng.IntN(m.InitialLifespan)
	m.Fatigue, m.Stress, m.Fame = rng.IntN(101), rng.IntN(101), rng.IntN(121)
	m.Form = rng.IntN(201) - 100
	m.Spoil, m.Fear = rng.IntN(101), rng.IntN(101)
	m.PlayType, m.LifeStage, m.LifeType, m.Rank, m.ArenaSpeed = rng.IntN(3), rng.IntN(10), rng.IntN(5), rng.IntN(6), rng.IntN(5)
	m.GutsRate = 5 + rng.IntN(16)
	m.TrainBoost = rng.IntN(1 << 16)
	m.Specials = rng.IntN(1 << 10)
	m.CupJellies = rng.IntN(40)
	m.GoldPeach, m.SilverPeach, m.ItemUsed = flag(), flag(), flag()
	for d := range m.Motivation {
		m.Motivation[d] = rng.IntN(101)
	}
	m.ItemLike, m.ItemDislike = rng.IntN(item.EmptyFrom), rng.IntN(item.EmptyFrom)
	m.Drug, m.DrugWeeks = 0, 0
	if rng.IntN(3) == 0 {
		m.Drug, m.DrugWeeks = 34+rng.IntN(8), 1+rng.IntN(4)
	}
	m.PrizeMoney = rng.IntN(100000)
	m.Record = monster.Record{
		Wins: rng.IntN(60), Losses: rng.IntN(40), KOs: rng.IntN(30),
		Injuries: rng.IntN(5), Runaways: rng.IntN(5), Sicknesses: rng.IntN(8), Resistance: rng.IntN(20),
	}

	for slot := range m.Learned {
		m.Learned[slot], m.Used[slot] = flag(), rng.IntN(100)
	}
	for column := range m.Active {
		m.Active[column] = monster.NoActiveMove
		if rng.IntN(5) != 0 {
			m.Active[column] = column*domain.Rows + rng.IntN(domain.Rows)
		}
	}

	p.Money, p.NextSale = rng.IntN(20000), rng.IntN(13)
	p.ErrantryCD = 255
	if flag() {
		p.ErrantryCD = rng.IntN(255)
	}
	for slot := range p.Items {
		p.Items[slot] = item.EmptyFrom
		if rng.IntN(3) != 0 {
			p.Items[slot] = rng.IntN(item.EmptyFrom)
		}
	}
	p.Ranch = player.Ranch{
		Known: true, Date: rng.IntN(8 * 48), Tournament: 1000 + rng.IntN(4), Grade: rng.IntN(11),
		House: rng.IntN(player.MaxLevel(player.House) + 1), Barn: rng.IntN(player.MaxLevel(player.Barn) + 1),
		Altar:  rng.IntN(player.MaxLevel(player.Altar) + 1),
		Energy: rng.IntN(300) - 10, EnergyKnown: true,
	}.WalkedTo(0, 0)
	for breed := range p.Ranch.Unlocked {
		p.Ranch.Unlocked[breed] = flag()
	}
	p.Breeder = "Demo Breeder"
	p.Input = rng.IntN(1 << 16)
}
