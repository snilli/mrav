package testutil

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"orm"
)

// ErrNoShippedDB is that no data/mrav.sqlite lies above the working directory.
var ErrNoShippedDB = errors.New("no data/mrav.sqlite above the working directory")

// OpenShipped opens the database the project ships, data/mrav.sqlite, which it finds by looking up from the
// directory of the test. Close it when done.
func OpenShipped(ctx context.Context) (*orm.Client, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	for {
		path := filepath.Join(dir, "data", "mrav.sqlite")
		if _, err := os.Stat(path); err == nil {
			return orm.Open(ctx, path)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ErrNoShippedDB
		}
		dir = parent
	}
}
