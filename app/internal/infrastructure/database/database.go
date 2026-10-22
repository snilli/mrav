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

func Open(ctx context.Context, path string) (*orm.Client, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("no database at %s: %w", path, err)
	}
	return orm.Open(ctx, path)
}

func resolve(path string, built []byte, dir string) (string, error) {
	if _, err := os.Stat(path); err == nil || len(built) == 0 {
		return path, nil
	}
	sum := sha256.Sum256(built)
	file := filepath.Join(dir, fmt.Sprintf("mrav-%x.sqlite", sum[:6]))
	if hasSize(file, len(built)) {
		return file, nil
	}
	if err := writeFile(file, built); err != nil && !hasSize(file, len(built)) {
		return "", fmt.Errorf("failed to write the built-in database: %w", err)
	}
	return file, nil
}

func hasSize(file string, size int) bool {
	info, err := os.Stat(file)
	return err == nil && info.Size() == int64(size)
}

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

func builtInDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	return filepath.Join(base, "mrav")
}

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
