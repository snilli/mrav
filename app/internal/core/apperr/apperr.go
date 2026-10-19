// Package apperr is the error vocabulary every bounded context shares: what went wrong, in words an
// outer layer can act on without knowing the domain. Pure Go.
package apperr

import "errors"

var (
	ErrInvalid  = errors.New("invalid")   // the input breaks a rule of the catalog
	ErrNotFound = errors.New("not found") // no such thing
)

// Error names the field at fault and how to fix it; Kind is one of the Err values above.
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
