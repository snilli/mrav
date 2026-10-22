package testutil

import (
	"context"
	"fmt"
	"orm"
)

func NewClient(ctx context.Context) (*orm.Client, error) {
	client, err := orm.OpenMemory(ctx)
	if err != nil {
		return nil, err
	}
	if err := client.Schema.Create(ctx); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return client, nil
}
