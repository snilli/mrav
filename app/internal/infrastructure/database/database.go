// Package database opens the SQLite file the viewer reads.
package database

import (
	"context"
	"fmt"
	"orm"
	"os"

	"github.com/rs/zerolog/log"
	"go.uber.org/fx"

	"app/config"
)

// Open opens a database that exists. Unlike orm.Open it never creates one: a viewer pointed at a missing
// file should say so, not show an empty window.
func Open(ctx context.Context, path string) (*orm.Client, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("no database at %s: %w", path, err)
	}
	return orm.Open(ctx, path)
}

// ProvideDatabase opens the database the config names and closes it with the app. It never migrates.
func ProvideDatabase(lc fx.Lifecycle, cfg *config.Config) (*orm.Client, error) {
	client, err := Open(context.Background(), cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the database: %w", err)
	}
	log.Info().Str("path", cfg.DBPath).Msg("Opened the move database")
	lc.Append(fx.Hook{OnStop: func(context.Context) error { return client.Close() }})
	return client, nil
}
