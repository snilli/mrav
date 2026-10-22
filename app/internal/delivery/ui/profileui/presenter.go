package profileui

import (
	"fmt"
	"strconv"
	"strings"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile/domain"
)

type Tone int

const (
	ToneNone Tone = iota
	ToneRed
	ToneOrangeRed
	ToneOrange
	ToneYellow
	ToneYellowGreen
	ToneLimeGreen
	ToneLightGreen
	ToneLightYellow
	ToneLightPink
	ToneHotPink
	ToneGray
	ToneBlack
	ToneAccent
)

type Box struct {
	Text string
	Tone Tone
}

type Options struct {
	GutsPerSecond  bool
	GrowthFromZero bool
}

type Bar struct {
	Ratio float64
	Tone  Tone
}

type Bars struct {
	Stats     [monster.StatCount]float64
	Fatigue   Bar
	Stress    Bar
	Nature    Bar
	Lifespan  Bar
	LifeIndex Bar
}

type CocoonView struct {
	CupJellies Box
	Ready      bool
}

type MonsterView struct {
	Breed           Box
	GivenName       string
	MainGenus       string
	SubGenus        string
	Age             string
	LifeStage       string
	LifeType        string
	Lifespan        string
	InitialLifespan string
	Stats           [monster.StatCount]string
	StatTotal       string
	Growth          [monster.StatCount]Box
	Loyalty         string
	Spoil           string
	Fear            string
	Fame            Box
	Rank            string
	Stress          string
	Fatigue         string
	LifeIndex       Box
	Form            string
	NatureBase      string
	Nature          Box
	GutsRate        Box
	ArenaSpeed      Box
	Specials        Box
	Like            Box
	Dislike         Box
	PlayType        string
	Drug            string
	DrugWeeks       string
	DrugTip         string
	GoldPeach       bool
	SilverPeach     bool
	ItemUsed        bool
	Money           string
	PrizeMoney      string
	NextSale        string
	ErrantrySale    bool
	Cocoon          *CocoonView
	Bars            Bars
	FatigueWord     string
	Record          string
	History         string
	Ranch           RanchView
	Expeditions     ExpeditionsView
}

var gradeTones = [...]Tone{ToneOrangeRed, ToneOrange, ToneYellow, ToneYellowGreen, ToneLimeGreen}

func arenaTone(speed int) Tone {
	if speed >= 0 && speed < len(gradeTones) {
		return gradeTones[speed]
	}
	return ToneNone
}

func gradeTone(rate int) Tone {
	if rate >= len(gradeTones) && rate <= 10 {
		return ToneHotPink
	}
	return arenaTone(rate)
}

func NewMonsterView(p domain.Profile, o Options) MonsterView {
	m := p.Monster
	v := MonsterView{
		Breed:           Box{Text: p.Breed.Name},
		GivenName:       m.Name,
		MainGenus:       p.MainGenus,
		SubGenus:        p.SubGenus,
		Age:             "---",
		LifeStage:       domain.LifeStageText(m.LifeStage, m.Lifespan),
		LifeType:        domain.LifeTypeName(m.LifeType),
		Lifespan:        strconv.Itoa(m.Lifespan) + "w",
		InitialLifespan: strconv.Itoa(m.InitialLifespan) + "w",
		Loyalty:         strconv.Itoa(domain.Loyalty(m.Spoil, m.Fear)),
		Spoil:           strconv.Itoa(m.Spoil),
		Fear:            strconv.Itoa(m.Fear),
		Fame:            Box{Text: strconv.Itoa(m.Fame)},
		Rank:            domain.RankLetter(m.Rank),
		Stress:          strconv.Itoa(m.Stress),
		Fatigue:         strconv.Itoa(m.Fatigue),
		LifeIndex:       lifeIndexBox(m),
		Form:            fmt.Sprintf("%d (%s)", m.Form, domain.FormLabel(m.Form)),
		NatureBase:      strconv.Itoa(m.NatureBase),
		Nature:          natureBox(m.Nature),
		GutsRate:        Box{Text: domain.GutsText(m.GutsRate, o.GutsPerSecond)},
		ArenaSpeed:      Box{Text: domain.ArenaGrade(m.ArenaSpeed), Tone: arenaTone(m.ArenaSpeed)},
		Like:            desireBox(p.Like, ToneLightGreen),
		Dislike:         desireBox(p.Dislike, ToneOrangeRed),
		PlayType:        domain.PlayTypeName(m.PlayType),
		Drug:            p.Drug,
		GoldPeach:       m.GoldPeach,
		SilverPeach:     m.SilverPeach,
		ItemUsed:        m.ItemUsed,
		Money:           strconv.Itoa(p.Player.Money) + "G",
		PrizeMoney:      strconv.Itoa(m.PrizeMoney) + "G",
		NextSale:        strconv.Itoa(p.Player.NextSale) + "w",
		ErrantrySale:    p.Player.ErrantrySaleReady(),
		FatigueWord:     domain.FatigueWord(m.Fatigue),
		Record:          fmt.Sprintf("%d W · %d L · %d KO", m.Record.Wins, m.Record.Losses, m.Record.KOs),
		History: fmt.Sprintf("sick %d · injured %d · ran away %d · resistance %d",
			m.Record.Sicknesses, m.Record.Injuries, m.Record.Runaways, m.Record.Resistance),
		Ranch:       newRanchView(p),
		Expeditions: newExpeditionsView(p),
	}
	if strings.Contains(p.Breed.Name, "[E]") || strings.Contains(p.Breed.Name, "(N/S)") {
		v.Breed.Tone = ToneLightPink
	}
	if m.Age != -1 {
		v.Age = fmt.Sprintf("%s (%dw)", domain.AgeText(m.Age), m.Age)
	}
	if m.Fame > 100 {
		v.Fame.Tone = ToneHotPink
	}
	if !p.Breed.GutsOK {
		v.GutsRate.Tone = ToneHotPink
	}
	specials := domain.SpecialsOf(m.Specials, m.Nature, int(m.Genus), int(m.SubGenus))
	v.Specials = Box{Text: specials.Text()}
	if specials.Invalid {
		v.Specials.Tone = ToneHotPink
	}

	v.Bars = Bars{
		Fatigue:   Bar{Ratio: fraction(m.Fatigue, 100), Tone: conditionTone(m.Fatigue)},
		Stress:    Bar{Ratio: fraction(m.Stress, 100), Tone: conditionTone(m.Stress)},
		Nature:    Bar{Ratio: fraction(m.Nature+100, 200)},
		Lifespan:  Bar{Ratio: fraction(m.Lifespan, m.InitialLifespan), Tone: lifespanTone(m.Lifespan, m.InitialLifespan)},
		LifeIndex: Bar{Ratio: fraction(domain.LifeIndexOf(m.Fatigue, m.Stress).Value, 280), Tone: v.LifeIndex.Tone},
	}
	total := 0
	for s, stat := range m.Stats {
		v.Bars.Stats[s] = fraction(stat, 999)
		v.Stats[s] = strconv.Itoa(stat)
		total += stat
		rate := m.Growth[s]
		number := rate + 1
		if o.GrowthFromZero {
			number = rate
		}
		v.Growth[s] = Box{Text: fmt.Sprintf("%s / %d", domain.GrowthGrade(rate), number), Tone: gradeTone(rate)}
	}
	v.StatTotal = strconv.Itoa(total)
	v.Stats[monster.Speed] += fmt.Sprintf(" [%d]", domain.EffectiveSpeed(m.Stats[monster.Speed], m.Form))
	v.Stats[monster.Defense] += fmt.Sprintf(" [%d]", domain.EffectiveDefense(m.Stats[monster.Defense], m.Form))

	if domain.IsDrug(m.Drug) {
		v.DrugWeeks = fmt.Sprintf("(%dw)", m.DrugWeeks)
		v.DrugTip = domain.DrugTip(m.Drug)
	}
	if domain.ShowsCocoon(m) {
		jellies := Box{Text: strconv.Itoa(m.CupJellies)}
		if m.CupJellies >= 30 {
			jellies.Tone = ToneLightGreen
		}
		v.Cocoon = &CocoonView{CupJellies: jellies, Ready: domain.CocoonReady(m)}
	}
	return v
}

func fraction(part, whole int) float64 {
	if whole <= 0 {
		return 0
	}
	return min(max(float64(part)/float64(whole), 0), 1)
}

func conditionTone(value int) Tone {
	switch {
	case value < 40:
		return ToneLimeGreen
	case value < 70:
		return ToneOrange
	}
	return ToneRed
}

func lifespanTone(left, initial int) Tone {
	switch ratio := fraction(left, initial); {
	case ratio > 0.5:
		return ToneLimeGreen
	case ratio > 0.25:
		return ToneOrange
	}
	return ToneRed
}

func natureBox(effective int) Box {
	label := domain.NatureOf(effective)
	b := Box{Text: fmt.Sprintf("%d (%s)", effective, label)}
	switch label {
	case domain.Bad:
		b.Tone = ToneGray
	case domain.Worst:
		b.Tone = ToneBlack
	case domain.Good:
		b.Tone = ToneLightYellow
	}
	return b
}

func desireBox(name string, tone Tone) Box {
	b := Box{Text: name}
	if name == "Tablet" || name == "Battle" {
		b.Tone = tone
	}
	return b
}

func MotivationTone(percent int) Tone {
	switch domain.MotivationOf(percent) {
	case domain.LevelLow:
		return ToneOrangeRed
	case domain.LevelMid:
		return ToneLightYellow
	case domain.LevelHigh:
		return ToneLightGreen
	}
	return ToneNone
}

func lifeIndexBox(m monster.Monster) Box {
	index := domain.LifeIndexOf(m.Fatigue, m.Stress)
	weeks := "Weeks"
	if index.WeeksLost == 1 {
		weeks = "Week"
	}
	tone := ToneLightGreen
	switch {
	case index.WeeksLost >= 7:
		tone = ToneRed
	case index.WeeksLost >= 5:
		tone = ToneOrangeRed
	case index.WeeksLost >= 3:
		tone = ToneOrange
	case index.WeeksLost >= 1:
		tone = ToneLightYellow
	}
	return Box{Text: fmt.Sprintf("%d (-%d %s)", index.Value, index.WeeksLost, weeks), Tone: tone}
}
