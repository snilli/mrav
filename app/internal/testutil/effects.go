package testutil

import (
	"context"

	"app/internal/core/planner/domain"
	"app/internal/core/planner/repository/db/plannerrepo"
)

func ShippedEffects(ctx context.Context) (domain.Effects, error) {
	client, err := OpenShipped(ctx)
	if err != nil {
		return domain.Effects{}, err
	}
	defer client.Close()
	list, err := plannerrepo.New(client).ListEffects(ctx)
	if err != nil {
		return domain.Effects{}, err
	}
	return domain.NewEffects(list), nil
}
