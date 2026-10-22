package testutil

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"orm"
)

var ErrNoShippedDB = errors.New("no data/mrav.sqlite above the working directory")

func OpenShipped(ctx context.Context) (*orm.Client, error) {
	return OpenShippedTraced(ctx, nil)
}

func OpenShippedTraced(ctx context.Context, trace func(query string, args []any)) (*orm.Client, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	for {
		path := filepath.Join(dir, "data", "mrav.sqlite")
		if _, err := os.Stat(path); err == nil {
			if trace == nil {
				return orm.Open(ctx, path)
			}
			return orm.OpenTraced(ctx, path, trace)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ErrNoShippedDB
		}
		dir = parent
	}
}
