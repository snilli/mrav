package status

import "reflect"

type Seen[T any] struct {
	shown bool
	value T
	err   string
}

func (s *Seen[T]) Changed(v T, err error) bool {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	if s.shown && msg == s.err && reflect.DeepEqual(v, s.value) {
		return false
	}
	s.shown, s.value, s.err = true, v, msg
	return true
}
