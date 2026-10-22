package profile

import (
	"context"

	"app/internal/core/profile/domain"
)

type ProfileService interface {
	Current(ctx context.Context) (domain.Profile, error)
}
