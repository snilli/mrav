package status

import "reflect"

// Seen is what a page was last shown, so that a load that brings the same again draws nothing, as React.memo does: the
// pages are loaded twice a second, and what they show changes far less often. The zero value has seen nothing.
type Seen[T any] struct {
	shown bool
	value T
	err   string
}

// Changed says whether v and err are not what was last shown, and keeps them when they are not. Errors are the same when
// they say the same.
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
