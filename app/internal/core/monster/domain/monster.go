package domain

const (
	SlotCount    = 24
	ActiveGroups = 4
	NoActiveMove = 24
	DrillCount   = 10
)

type Genus int

type Stat int

const (
	Life Stat = iota
	Pow
	Int
	Skill
	Speed
	Defense
	StatCount = iota
)

type Monster struct {
	Genus    Genus
	SubGenus Genus
	Stats    [StatCount]int
	Nature   int
	Learned  [SlotCount]bool
	Used     [SlotCount]int
	Active   [ActiveGroups]int

	Name            string
	Age             int
	Lifespan        int
	InitialLifespan int
	NatureBase      int
	Fatigue         int
	Stress          int
	Fame            int
	Form            int
	Spoil           int
	Fear            int
	Growth          [StatCount]int
	TrainBoost      int
	CupJellies      int
	GoldPeach       bool
	SilverPeach     bool
	ItemUsed        bool
	Motivation      [DrillCount]int
	PatchedSwim     bool
	PlayType        int
	Drug            int
	DrugWeeks       int
	ItemLike        int
	ItemDislike     int
	Rank            int
	LifeStage       int
	LifeType        int
	ArenaSpeed      int
	GutsRate        int
	Specials        int
	PrizeMoney      int
	Record          Record
}

type Record struct {
	Wins, Losses, KOs                          int
	Injuries, Runaways, Sicknesses, Resistance int
}

func (m Monster) StatSum(stats ...Stat) int {
	sum := 0
	for _, s := range stats {
		if s >= 0 && s < StatCount {
			sum += m.Stats[s]
		}
	}
	return sum
}
