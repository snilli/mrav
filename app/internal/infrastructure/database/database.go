// Package database opens the SQLite file the viewer reads.
package database

import (
	"context"
	"crypto/sha256"
	"fmt"
	"orm"
	"os"
	"path/filepath"

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

// resolve is the file to open: the one at path when there is one, otherwise the database built into the program,
// written into dir, otherwise path again, so that Open says what is missing. The built-in file is named after
// its contents and left alone once it is there, because another copy of the viewer may have it open.
func resolve(path string, built []byte, dir string) (string, error) {
	if _, err := os.Stat(path); err == nil || len(built) == 0 {
		return path, nil
	}
	sum := sha256.Sum256(built)
	file := filepath.Join(dir, fmt.Sprintf("mrav-%x.sqlite", sum[:6]))
	if hasSize(file, len(built)) {
		return file, nil
	}
	// A copy that started at the same moment may have won the race to write it, which is as good.
	if err := writeFile(file, built); err != nil && !hasSize(file, len(built)) {
		return "", fmt.Errorf("failed to write the built-in database: %w", err)
	}
	return file, nil
}

func hasSize(file string, size int) bool {
	info, err := os.Stat(file)
	return err == nil && info.Size() == int64(size)
}

// writeFile writes through a temporary file, so that a write that is cut short never leaves a file under the name.
func writeFile(file string, data []byte) error {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "mrav-*.tmp")
	if err != nil {
		return err
	}
	_, err = tmp.Write(data)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), file)
}

// builtInDir is where the built-in database is written: the user's cache, or the temporary files without one.
func builtInDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "mrav")
}

// ProvideDatabase opens the database the config names, or the one built into the program when there is no file
// there, and closes it with the app. It never migrates.
func ProvideDatabase(lc fx.Lifecycle, cfg *config.Config) (*orm.Client, error) {
	path, err := resolve(cfg.DBPath, embedded, builtInDir())
	if err != nil {
		return nil, err
	}
	client, err := Open(context.Background(), path)
	if err != nil {
		return nil, fmt.Errorf("failed to open the database: %w", err)
	}
	if path != cfg.DBPath {
		log.Info().Str("wanted", cfg.DBPath).Msg("No database there, so using the one built in")
	}
	log.Info().Str("path", path).Msg("Opened the move database")
	lc.Append(fx.Hook{OnStop: func(context.Context) error { return client.Close() }})
	return client, nil
}
