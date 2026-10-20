package domain

import "fmt"

// The game keeps its calendar in weeks: four to a month, twelve months to a year.
const (
	WeeksPerMonth = 4
	MonthsPerYear = 12
	weeksPerYear  = WeeksPerMonth * MonthsPerYear
)

// The game's counter of weeks begins in the year firstYear, in the week of that year startWeek, counted from
// January's first week as 0. Neither is read from the game: both were worked out from one date seen in it, 1018,
// March, week 1, so the date shown comes with the counter itself, to be checked against the game's own.
const (
	firstYear = 1000
	startWeek = 12 // April, week 1
)

// The gold that has to be on hand beyond the cost of an upgrade for the game to offer it, and when it offers it.
const (
	PromptBuffer = 5000
	UpgradeMonth = 5 // May
	UpgradeWeek  = 4
)

// MonthNames are the months of the calendar, abbreviated.
var MonthNames = [MonthsPerYear]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// Ranch is what the game keeps of the ranch and its calendar. Only the Steam release is known to keep it where this
// viewer can read it, so Known says whether the rest is real.
type Ranch struct {
	Known      bool
	Date       int              // weeks since the game began
	Tournament int              // the year of the first IMA-FIMBA tournament; 0 where it was not read
	Grade      int              // the breeder's grade: 0 for Beginner, 1 to 9, and 10 for Master
	House      int              // how many upgrades have been made: 0 to 4
	Barn       int              // the stable: 0 to 2
	Altar      int              // 0 for no altar, then 1 to 3
	Unlocked   [BreedCount]bool // which main breeds the game has unlocked, by their codes

	// Energy is what the game last kept as an expedition's energy, which it can take below nothing. Whether it is
	// made of an expedition now is not known: the value stays after the expedition ends. EnergyKnown is false where
	// the viewer could not read it.
	Energy      int
	EnergyKnown bool

	// Tile is the tile of an expedition's map that its monster last stood on, and Steps how many it had walked. Like
	// the energy they are kept after the expedition, and WalkKnown is false where the viewer could not read them.
	Tile, Steps int
	WalkKnown   bool
}

// The game counts the tiles of an expedition's map in rows of MapWidth, and every trip starts on the tile of column
// StartColumn of row StartRow of that count: the maps are laid on it with their start there.
const (
	MapWidth    = 32
	StartColumn = 15
	StartRow    = 30
)

// Walked is where the monster is as the game shows it on a trip: how far to the right of the start it is, and how
// far up. Either can be below 0.
func (r Ranch) Walked() (x, y int) {
	return r.Tile%MapWidth - StartColumn, StartRow - r.Tile/MapWidth
}

// WalkedTo is the ranch with the monster put x tiles right of the start of the trip and y tiles up, as Walked says it, and
// known to be there.
func (r Ranch) WalkedTo(x, y int) Ranch {
	r.Tile = (StartRow-y)*MapWidth + StartColumn + x
	r.WalkKnown = true
	return r
}

// BreedCount is how many main breeds the game has, Pixie to Naga.
const BreedCount = 38

// Date is a week of the calendar: the year, the month from 1 and the week of the month from 1.
type Date struct{ Year, Month, Week int }

func (d Date) String() string {
	return fmt.Sprintf("Y%d %s wk %d", d.Year, MonthNames[d.Month-1], d.Week)
}

// Calendar is the week the game is in now.
func (r Ranch) Calendar() Date {
	w := r.Date + startWeek
	return Date{
		Year:  firstYear + w/weeksPerYear,
		Month: w%weeksPerYear/WeeksPerMonth + 1,
		Week:  w%WeeksPerMonth + 1,
	}
}

// InWeek is the ranch with its date moved to a month and a week, from 1, of the year that the date is in, or of the next
// year where the game had not begun by then. A ranch that is moved is known.
func (r Ranch) InWeek(month, week int) Ranch {
	date := (r.Date+startWeek)/weeksPerYear*weeksPerYear + (month-1)*WeeksPerMonth + week - 1 - startWeek
	if date < 0 {
		date += weeksPerYear
	}
	r.Known, r.Date = true, date
	return r
}

// WeeksUntil is how many weeks there are from now until the calendar next reaches a week of the year, 0 when it is
// this week.
func (r Ranch) WeeksUntil(month, week int) int {
	now := (r.Date + startWeek) % weeksPerYear
	target := (month-1)*WeeksPerMonth + week - 1
	return (target - now + weeksPerYear) % weeksPerYear
}

// The IMA-FIMBA tournament is held in July, week 1, every fourth year from the year of the first.
const (
	TournamentMonth = 7
	TournamentWeek  = 1
	tournamentEvery = 4
)

// NextTournament is the tournament the calendar reaches next, and how many weeks away it is, 0 on the day. ok is
// false where the year of the first was not read.
func (r Ranch) NextTournament() (when Date, weeks int, ok bool) {
	if !r.Known || r.Tournament == 0 {
		return Date{}, 0, false
	}
	now := r.Calendar()
	inYear := (r.Date + startWeek) % weeksPerYear
	day := (TournamentMonth-1)*WeeksPerMonth + TournamentWeek - 1
	for year := max(now.Year, r.Tournament); ; year++ {
		if (year-r.Tournament)%tournamentEvery != 0 {
			continue
		}
		if weeks = (year-now.Year)*weeksPerYear + day - inYear; weeks >= 0 {
			return Date{Year: year, Month: TournamentMonth, Week: TournamentWeek}, weeks, true
		}
	}
}

// Facility is a part of the ranch that can be upgraded.
type Facility int

const (
	House Facility = iota
	Barn
	Altar
)

// Upgrade is one step up for a facility: what it costs, and the breeder grade it asks for. Needed is the gold on hand
// that a guide gives for it where a guide gives one, which is not always the cost and 5,000; 0 where none does.
type Upgrade struct{ Cost, Grade, Needed int }

// GoldNeeded is the gold on hand that the game wants before it offers the upgrade.
func (u Upgrade) GoldNeeded() int {
	if u.Needed > 0 {
		return u.Needed
	}
	return u.Cost + PromptBuffer
}

// The altar is not offered with the house and the stable. Its upgrades come in March, week 3, once 140 weeks have
// passed since the funeral that made it, or since its last upgrade.
const (
	AltarMonth = 3
	AltarWeek  = 3
	AltarWait  = 140
)

// UpgradeTime is the month and the week that the game offers a facility's upgrades in.
func UpgradeTime(f Facility) (month, week int) {
	if f == Altar {
		return AltarMonth, AltarWeek
	}
	return UpgradeMonth, UpgradeWeek
}

// steps are the upgrades of each facility, from the first. The altar's are from the altar it starts with, which is
// level 1.
var steps = [...][]Upgrade{
	House: {{Cost: 5000, Grade: 3}, {Cost: 8000, Grade: 5}, {Cost: 12000, Grade: 6}, {Cost: 20000, Grade: 8}},
	Barn:  {{Cost: 15000, Grade: 4}, {Cost: 30000, Grade: 7}},
	Altar: {{Cost: 5000, Needed: 8000}, {Cost: 8000, Needed: 13000}},
}

// MaxLevel is the level a facility ends at.
func MaxLevel(f Facility) int {
	if f == Altar {
		return len(steps[Altar]) + 1
	}
	return len(steps[f])
}

// Level is how far a facility has been upgraded.
func (r Ranch) Level(f Facility) int {
	switch f {
	case House:
		return r.House
	case Barn:
		return r.Barn
	}
	return r.Altar
}

// Next is the upgrade a facility would have next. It is false when there is none: the facility is at its top, or
// it is an altar the ranch does not have.
func (r Ranch) Next(f Facility) (Upgrade, bool) {
	i := r.Level(f)
	if f == Altar {
		i-- // the altar's steps start from level 1
	}
	if i < 0 || i >= len(steps[f]) {
		return Upgrade{}, false
	}
	return steps[f][i], true
}

// Slots is how many places for items the house gives: 12, and 2 more for each upgrade, up to the game's 20. When
// the house is not known they are all there.
func (r Ranch) Slots() int {
	if !r.Known {
		return ItemSlots
	}
	return min(12+2*r.House, ItemSlots)
}
