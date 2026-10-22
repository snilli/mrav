package profileui

import (
	"fmt"
	"strconv"
	"strings"

	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
)

type RanchView struct {
	Known       bool
	Date        string
	UpgradeDay  string
	Tournament  string
	Facilities  []FacilityView
	Breeds      []BreedView
	BreedsTitle string
}

type BreedView struct {
	Name     string
	Unlocked bool
}

type FacilityView struct {
	Name       string
	Level, Max int
	Gives      string
	Next       string
	State      Box
	Offer      string
}

var facilityNames = [...]string{player.House: "House", player.Barn: "Stable", player.Altar: "Altar"}

func newRanchView(p domain.Profile) RanchView {
	r := p.Player.Ranch
	if !r.Known {
		return RanchView{}
	}
	v := RanchView{Known: true, Date: fmt.Sprintf("%s (week %d)", r.Calendar(), r.Date), UpgradeDay: upgradeDay(r)}
	if when, weeks, ok := r.NextTournament(); ok {
		v.Tournament = when.String() + ", " + inWeeks(weeks)
	}
	for _, f := range []player.Facility{player.House, player.Barn, player.Altar} {
		v.Facilities = append(v.Facilities, newFacilityView(f, p.Player))
	}
	unlocked := 0
	for _, b := range p.Breeds {
		v.Breeds = append(v.Breeds, BreedView{Name: b.Name, Unlocked: b.Unlocked})
		if b.Unlocked {
			unlocked++
		}
	}
	if len(v.Breeds) > 0 {
		v.BreedsTitle = fmt.Sprintf("Breeds unlocked %d/%d", unlocked, len(v.Breeds))
	}
	return v
}

func upgradeDay(r player.Ranch) string {
	day := fmt.Sprintf("%s wk %d", player.MonthNames[player.UpgradeMonth-1], player.UpgradeWeek)
	return day + ", " + inWeeks(r.WeeksUntil(player.UpgradeMonth, player.UpgradeWeek))
}

func inWeeks(weeks int) string {
	switch weeks {
	case 0:
		return "this week"
	case 1:
		return "in 1 week"
	}
	return fmt.Sprintf("in %d weeks", weeks)
}

func newFacilityView(f player.Facility, p player.Player) FacilityView {
	r := p.Ranch
	level, top := r.Level(f), player.MaxLevel(f)
	v := FacilityView{Name: facilityNames[f], Level: level, Max: top}
	if f == player.House {
		v.Gives = fmt.Sprintf("%d places", r.Slots())
	}
	switch up, ok := r.Next(f); {
	case ok:
		v.Next = "next " + Gold(up.Cost)
		if f == player.Altar {
			month, week := player.UpgradeTime(f)
			v.Offer = fmt.Sprintf("Offered %s wk %d, %d weeks after the last step", player.MonthNames[month-1], week, player.AltarWait)
		}
		v.State = upgradeState(up, r.Grade, p.Money)
		if f == player.Altar && v.State.Text == "Ready" {
			v.State = Box{Text: "Gold is enough", Tone: ToneLightGreen}
		}
	case f == player.Altar && level == 0:
		v.State = Box{Text: "No altar yet", Tone: ToneGray}
	default:
		v.State = Box{Text: "Fully upgraded", Tone: ToneGray}
	}
	return v
}

func upgradeState(up player.Upgrade, grade, money int) Box {
	var missing []string
	if grade < up.Grade {
		missing = append(missing, fmt.Sprintf("breeder grade %d (now %d)", up.Grade, grade))
	}
	if need := up.GoldNeeded(); money < need {
		missing = append(missing, Gold(need-money)+" more")
	}
	if len(missing) == 0 {
		return Box{Text: "Ready", Tone: ToneLimeGreen}
	}
	return Box{Text: needs(missing), Tone: ToneOrange}
}

func needs(items []string) string {
	if len(items) < 2 {
		return "Needs " + strings.Join(items, "")
	}
	return "Needs " + strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}

func Gold(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s + " G"
}
