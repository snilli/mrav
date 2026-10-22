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

	DBPath string `env:"DB_PATH" default:"../data/mrav.sqlite" validate:"required"`

	RefreshMS int `env:"REFRESH_MS" default:"500" validate:"min=50"`

	Game string `env:"GAME" default:"auto" validate:"oneof=auto demo duckstation mr2dx"`
}

var validate = validator.New()

func Load() (*Config, error) {
	_ = godotenv.Load("../.env")
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
