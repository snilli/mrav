package domain

type Food int

const (
	Potato Food = iota
	Fish
	Meat
	Milk
	CupJelly
	Tablet

	NoFood Food = -1
)

var (
	foodKeys  = [...]string{"potato", "fish", "meat", "milk", "cup_jelly", "tablet"}
	foodNames = [...]string{"Potato", "Fish", "Meat", "Milk", "Cup Jelly", "Tablet"}
)

func Foods() []Food { return []Food{Potato, Milk, Fish, CupJelly, Meat, Tablet} }

func (f Food) Name() string {
	if f == NoFood {
		return "Other"
	}
	return foodNames[f]
}

func (f Food) Key() string {
	if f == NoFood {
		return "none"
	}
	return foodKeys[f]
}

func ParseFood(key string) (Food, bool) {
	for f, k := range foodKeys {
		if k == key {
			return Food(f), true
		}
	}
	return NoFood, false
}

type Reaction int

const (
	Like Reaction = iota
	Neutral
	Dislike
)

var reactionKeys = [...]string{"like", "neutral", "dislike"}

func (r Reaction) Key() string { return reactionKeys[r] }

func ParseReaction(key string) (Reaction, bool) {
	for r, k := range reactionKeys {
		if k == key {
			return Reaction(r), true
		}
	}
	return Like, false
}

type Stage int

const (
	AnyStage Stage = iota
	Baby
	Grown
)

var stageKeys = [...]string{"any", "baby", "grown"}

func (s Stage) Key() string { return stageKeys[s] }

func ParseStage(key string) (Stage, bool) {
	for s, k := range stageKeys {
		if k == key {
			return Stage(s), true
		}
	}
	return AnyStage, false
}

type FoodReaction struct {
	Genus    int
	Food     Food
	Stage    Stage
	Reaction Reaction
}

type reactionKey struct {
	genus int
	food  Food
	stage Stage
}

type Reactions struct{ byKey map[reactionKey]Reaction }

func NewReactions(list []FoodReaction) Reactions {
	r := Reactions{byKey: make(map[reactionKey]Reaction, len(list))}
	for _, f := range list {
		r.byKey[reactionKey{f.Genus, f.Food, f.Stage}] = f.Reaction
	}
	return r
}

func (r Reactions) Of(genus int, food Food, lifeStage int) Reaction {
	if reaction, ok := r.byKey[reactionKey{genus, food, AnyStage}]; ok {
		return reaction
	}
	stage := Grown
	if lifeStage == 1 {
		stage = Baby
	}
	if reaction, ok := r.byKey[reactionKey{genus, food, stage}]; ok {
		return reaction
	}
	if food == Meat {
		return Dislike
	}
	return Like
}

var foodStress = [...][3]int{
	Potato:   {0, 4, 16},
	Fish:     {-6, -3, 2},
	Meat:     {-8, -6, 0},
	Milk:     {-3, -2, 4},
	CupJelly: {-7, -5, 1},
	Tablet:   {-15, -13, -10},
}

func FoodStress(food Food, reaction Reaction) int { return foodStress[food][reaction] }
