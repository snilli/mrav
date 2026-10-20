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
	Gives      string // what its level gives, where it gives something
	Next       string // what its next step costs; empty where there is none
	State      Box    // where that step stands, in words: ready, what it needs, or that there is nothing more to do
	Offer      string // when the game offers its steps, where that is a time of its own: the altar's
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
	v := FacilityView{Name: facilityNames[f], Level: level, Max: top}
	if f == player.House {
		v.Gives = fmt.Sprintf("%d places", r.Slots())
	}
	switch up, ok := r.Next(f); {
	case ok:
		v.Next = "next " + Gold(up.Cost)
		if f == player.Altar { // its own time: not the house's and the stable's May
			month, week := player.UpgradeTime(f)
			v.Offer = fmt.Sprintf("Offered %s wk %d, %d weeks after the last step", player.MonthNames[month-1], week, player.AltarWait)
		}
		v.State = upgradeState(up, r.Grade, p.Money)
		if f == player.Altar && v.State.Text == "Ready" {
			// the gold is all that is known: when it last had a step is not kept anywhere that is read
			v.State = Box{Text: "Gold is enough", Tone: ToneLightGreen}
		}
	case f == player.Altar && level == 0:
		v.State = Box{Text: "No altar yet", Tone: ToneGray}
	default:
		v.State = Box{Text: "Fully upgraded", Tone: ToneGray}
	}
	return v
}

// upgradeState says whether an upgrade can be had: ready when the breeder's grade is enough and the gold on hand is
// the cost and the game's 5,000 more, and otherwise in words what it needs.
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

// needs says what is lacking as a sentence: "Needs a", "Needs a and b", "Needs a, b and c".
func needs(items []string) string {
	if len(items) < 2 {
		return "Needs " + strings.Join(items, "")
	}
	return "Needs " + strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}

// Gold writes an amount of gold with a comma between thousands: 12,000 G.
func Gold(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s + " G"
}
