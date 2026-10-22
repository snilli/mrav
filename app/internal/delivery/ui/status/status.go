package status

import (
	"errors"

	"app/internal/core/apperr"
)

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
