// Package testutil holds helpers for tests that need a real database.
package testutil

import (
	"context"
	"fmt"
	"orm"
)

// NewClient is a private in-memory SQLite with the schema created. Close it when done.
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
