package domain

import (
	"cmp"
	"fmt"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
)

const (
	sculptureCode  = 6
	geminisPotCode = 7
)

var itemFoods = map[int]Food{0: Potato, 1: Fish, 2: Meat, 3: Milk, 4: CupJelly, 5: Tablet}

type Settings struct {
	Version     monster.Version
	Main, Sub   int
	Like        Food
	Dislike     Food
	LifeStage   int
	Form        int
	Pots        int
	Sculptures  int
	Fatigue     int
	Stress      int
	Lifespan    int
	GoldPeach   bool
	SilverPeach bool
	ItemUsed    bool
	Month, Week int
}

func DefaultSettings() Settings {
	return Settings{Version: monster.MR2DX, Main: 35, Sub: 35, Like: Milk, Dislike: Potato, LifeStage: 1}
}

func (s Settings) ate(item Item) bool {
	return item == GoldPeach && s.GoldPeach || item == SilverPeach && s.SilverPeach
}

func foodOfItem(code int) Food {
	if food, ok := itemFoods[code]; ok {
		return food
	}
	return NoFood
}

func SettingsOf(m monster.Monster, items [player.ItemSlots]int) Settings {
	s := Settings{
		Version: cmp.Or(m.Version, monster.MR2DX), Main: int(m.Genus), Sub: int(m.SubGenus),
		Like: foodOfItem(m.ItemLike), Dislike: foodOfItem(m.ItemDislike),
		LifeStage: min(max(m.LifeStage+1, 1), 10),
		Form:      m.Form, Fatigue: m.Fatigue, Stress: m.Stress, Lifespan: m.Lifespan,
		GoldPeach: m.GoldPeach, SilverPeach: m.SilverPeach, ItemUsed: m.ItemUsed,
	}
	for _, code := range items {
		switch code {
		case geminisPotCode:
			s.Pots++
		case sculptureCode:
			s.Sculptures++
		}
	}
	return s
}

func (s Settings) In(r player.Ranch) Settings {
	if r.Known {
		d := r.Calendar()
		s.Month, s.Week = d.Month, d.Week
	}
	return s
}

func (s Settings) MonthName(month int) string {
	if s.Month == 0 {
		return fmt.Sprintf("Month %d", month)
	}
	return player.MonthNames[(s.Month+month-2)%player.MonthsPerYear]
}

func (s Settings) firstWeek() int { return min(max(s.Week, 1), player.WeeksPerMonth) }
