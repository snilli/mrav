// Package status says in words why a page has nothing to show.
package status

import (
	"errors"

	"app/internal/core/apperr"
)

// Message is what a page shows in place of its content when loading it failed: what a not-found error says,
// such as that the game is not running, or No monster when it says nothing; any other error as an error. It is
// empty when nothing failed.
func Message(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, apperr.ErrNotFound) {
		var e *apperr.Error
		if errors.As(err, &e) && e.Msg != "" {
			return e.Msg
		}
		return "No monster"
	}
	return "Error: " + err.Error()
}
