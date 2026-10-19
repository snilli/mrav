package profileui

import (
	"fmt"
	"strconv"
	"strings"

	player "app/internal/core/player/domain"
	"app/internal/core/profile/domain"
)

// RanchView is what the ranch panel shows. It is empty when the game does not keep the ranch where this viewer can
// read it.
type RanchView struct {
	Known      bool
	Date       string // the date, with the game's own count of weeks beside it, to check it by
	UpgradeDay string // when the game offers the upgrades, and how long to wait for it
	Tournament string // when the next IMA-FIMBA tournament is, and how long to wait for it; empty where not known
	Facilities []FacilityView
	// The main breeds and whether each is unlocked, with the title of their list; none when the game does not say.
	Breeds      []BreedView
	BreedsTitle string
}

// BreedView is a main breed and whether the breeder has unlocked it.
type BreedView struct {
	Name     string
	Unlocked bool
}

// FacilityView is one part of the ranch that can be upgraded.
type FacilityView struct {
	Name       string
	Level, Max int
	Ratio      float64
	Detail     string // what its level gives, where it gives something
	Next       string // the next step and what it costs, or why there is none
	State      Box    // where that step stands: ready, what is missing, or the top
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

// upgradeDay says when the game next offers the upgrades, and how long that is.
func upgradeDay(r player.Ranch) string {
	day := fmt.Sprintf("%s wk %d", player.MonthNames[player.UpgradeMonth-1], player.UpgradeWeek)
	return day + ", " + inWeeks(r.WeeksUntil(player.UpgradeMonth, player.UpgradeWeek))
}

// inWeeks says how long to wait: this week, in 1 week, in 9 weeks.
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
	v := FacilityView{Name: facilityNames[f], Level: level, Max: top, Ratio: float64(level) / float64(top)}
	if f == player.House {
		v.Detail = fmt.Sprintf("%d places", r.Slots())
	}
	switch up, ok := r.Next(f); {
	case ok:
		v.Next = "next " + gold(up.Cost)
		if up.Grade > 0 {
			v.Next += fmt.Sprintf(" · grade %d", up.Grade)
		}
		if f == player.Altar { // its own time: not the house's and the stable's May
			month, week := player.UpgradeTime(f)
			v.Next += fmt.Sprintf(" · %s wk %d, %d weeks after the last step", player.MonthNames[month-1], week, player.AltarWait)
		}
		v.State = upgradeState(up, r.Grade, p.Money)
		if f == player.Altar && v.State.Text == "READY" {
			// the gold is all that is known: when it last had a step is not kept anywhere that is read
			v.State = Box{Text: "gold is enough", Tone: ToneLightGreen}
		}
	case f == player.Altar && level == 0:
		v.Detail, v.Next = "none", "no altar yet"
	default:
		v.Next, v.State = "fully upgraded", Box{Text: "MAX", Tone: ToneGray}
	}
	return v
}

// upgradeState says whether an upgrade can be had: ready when the breeder's grade is enough and the gold on hand is
// the cost and the game's 5,000 more, and otherwise what is missing.
func upgradeState(up player.Upgrade, grade, money int) Box {
	var missing []string
	if grade < up.Grade {
		missing = append(missing, fmt.Sprintf("grade %d", up.Grade))
	}
	if need := up.GoldNeeded(); money < need {
		missing = append(missing, gold(need-money)+" more")
	}
	if len(missing) == 0 {
		return Box{Text: "READY", Tone: ToneLimeGreen}
	}
	return Box{Text: strings.Join(missing, " · "), Tone: ToneOrange}
}

// gold writes an amount of gold with a comma between thousands: 12,000 G.
func gold(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s + " G"
}
