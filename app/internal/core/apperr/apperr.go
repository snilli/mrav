package apperr

import "errors"

var (
	ErrInvalid  = errors.New("invalid")
	ErrNotFound = errors.New("not found")
)

type Error struct {
	Kind       error
	Field, Msg string
}

func (e *Error) Error() string {
	if e.Field == "" {
		return e.Msg
	}
	return e.Field + ": " + e.Msg
}

func (e *Error) Unwrap() error { return e.Kind }

func Invalid(field, msg string) error  { return &Error{ErrInvalid, field, msg} }
func NotFound(field, msg string) error { return &Error{ErrNotFound, field, msg} }
