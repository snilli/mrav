package domain

import "fmt"

const (
	WeeksPerMonth = 4
	MonthsPerYear = 12
	weeksPerYear  = WeeksPerMonth * MonthsPerYear
)

const (
	firstYear = 1000
	startWeek = 12
)

const (
	PromptBuffer = 5000
	UpgradeMonth = 5
	UpgradeWeek  = 4
)

var MonthNames = [MonthsPerYear]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

type Ranch struct {
	Known      bool
	Date       int
	Tournament int
	Grade      int
	House      int
	Barn       int
	Altar      int
	Unlocked   [BreedCount]bool

	Energy      int
	EnergyKnown bool

	Tile      int
	WalkKnown bool

	OffsetX, OffsetY int
	OffsetKnown      bool

	PhoenixDone, PhoenixKnown bool
}

const (
	MapWidth    = 32
	StartColumn = 15
	StartRow    = 30
)

func (r Ranch) Walked() (x, y int) {
	return r.Tile%MapWidth - StartColumn, StartRow - r.Tile/MapWidth
}

func (r Ranch) OnTrip(col, row int) bool {
	return r.WalkKnown && r.OffsetKnown && r.OffsetX == r.Tile%MapWidth-col && r.OffsetY == row-r.Tile/MapWidth
}

func (r Ranch) WalkedTo(x, y int) Ranch {
	r.Tile = (StartRow-y)*MapWidth + StartColumn + x
	r.WalkKnown = true
	return r
}

const BreedCount = 38

type Date struct{ Year, Month, Week int }

func (d Date) String() string {
	return fmt.Sprintf("Y%d %s wk %d", d.Year, MonthNames[d.Month-1], d.Week)
}

func (r Ranch) Calendar() Date {
	w := r.Date + startWeek
	return Date{
		Year:  firstYear + w/weeksPerYear,
		Month: w%weeksPerYear/WeeksPerMonth + 1,
		Week:  w%WeeksPerMonth + 1,
	}
}

func (r Ranch) InWeek(month, week int) Ranch {
	date := (r.Date+startWeek)/weeksPerYear*weeksPerYear + (month-1)*WeeksPerMonth + week - 1 - startWeek
	if date < 0 {
		date += weeksPerYear
	}
	r.Known, r.Date = true, date
	return r
}

func (r Ranch) WeeksUntil(month, week int) int {
	now := (r.Date + startWeek) % weeksPerYear
	target := (month-1)*WeeksPerMonth + week - 1
	return (target - now + weeksPerYear) % weeksPerYear
}

const (
	TournamentMonth = 7
	TournamentWeek  = 1
	tournamentEvery = 4
)

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

type Facility int

const (
	House Facility = iota
	Barn
	Altar
)

type Upgrade struct{ Cost, Grade, Needed int }

func (u Upgrade) GoldNeeded() int {
	if u.Needed > 0 {
		return u.Needed
	}
	return u.Cost + PromptBuffer
}

const (
	AltarMonth = 3
	AltarWeek  = 3
	AltarWait  = 140
)

func UpgradeTime(f Facility) (month, week int) {
	if f == Altar {
		return AltarMonth, AltarWeek
	}
	return UpgradeMonth, UpgradeWeek
}

var steps = [...][]Upgrade{
	House: {{Cost: 5000, Grade: 3}, {Cost: 8000, Grade: 5}, {Cost: 12000, Grade: 6}, {Cost: 20000, Grade: 8}},
	Barn:  {{Cost: 15000, Grade: 4}, {Cost: 30000, Grade: 7}},
	Altar: {{Cost: 5000, Needed: 8000}, {Cost: 8000, Needed: 13000}},
}

func MaxLevel(f Facility) int {
	if f == Altar {
		return len(steps[Altar]) + 1
	}
	return len(steps[f])
}

func (r Ranch) Level(f Facility) int {
	switch f {
	case House:
		return r.House
	case Barn:
		return r.Barn
	}
	return r.Altar
}

func (r Ranch) Next(f Facility) (Upgrade, bool) {
	i := r.Level(f)
	if f == Altar {
		i--
	}
	if i < 0 || i >= len(steps[f]) {
		return Upgrade{}, false
	}
	return steps[f][i], true
}

func (r Ranch) Slots() int {
	if !r.Known {
		return ItemSlots
	}
	return min(12+2*r.House, ItemSlots)
}
