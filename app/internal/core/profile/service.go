package profile

import (
	"context"

	"app/internal/core/profile/domain"
)

// ProfileService is the Monster page's use cases, in service/profilesvc.
type ProfileService interface {
	// Current is the profile of the monster the game shows now. It is apperr.ErrNotFound when the game shows no
	// monster or cannot be reached.
	Current(ctx context.Context) (domain.Profile, error)
}
