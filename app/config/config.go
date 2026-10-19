package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go-simpler.org/env"
)

type Config struct {
	LogLevel string `env:"LOG_LEVEL" default:"info"`

	// DBPath is the SQLite file `import` writes and the viewer reads. Commands run from app/.
	DBPath string `env:"DB_PATH" default:"../data/mrav.sqlite" validate:"required"`

	// RefreshMS is how often the viewer asks for the monster, in milliseconds (MR2AV polls four times a second).
	RefreshMS int `env:"REFRESH_MS" default:"500" validate:"min=50"`

	// Game is where the monster comes from: "demo" is the panel on screen; "duckstation" and "mr2dx" read that game;
	// "auto" reads whichever is running, and is the demo panel where no game can be read.
	Game string `env:"GAME" default:"auto" validate:"oneof=auto demo duckstation mr2dx"`
}

var validate = validator.New()

func Load() (*Config, error) {
	_ = godotenv.Load("../.env") // the project root's .env, when run from app/
	_ = godotenv.Load()
	cfg := &Config{}
	if err := env.Load(cfg, nil); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	return cfg, nil
}

func (c *Config) SetupLogger() {
	level := zerolog.InfoLevel
	switch c.LogLevel {
	case "debug":
		level = zerolog.DebugLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"})
}
