package domain

func StageStarts(shares [10]int, lifespan int) [10]int {
	var starts [10]int
	starts[0] = lifespan
	left := 100
	for i, share := range shares[:9] {
		left -= share
		starts[i+1] = lifespan * left / 100
	}
	return starts
}
