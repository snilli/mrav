package gamememory

func attachFirst[P any](pids []P, attach func(P) (Memory, error)) (Memory, error) {
	var first error
	for _, pid := range pids {
		m, err := attach(pid)
		if err == nil {
			return m, nil
		}
		if first == nil {
			first = err
		}
	}
	return nil, first
}
