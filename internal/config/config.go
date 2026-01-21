package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	BotToken          string `env:"BOT_TOKEN"`
	DBDSN             string `env:"DB_DSN"`
	DefaultWeeklyNorm int32  `env:"DEFAULT_WEEKLY_NORM" envDefault:"100"`
}

func Load() (Config, error) {
	cfg := Config{}

	if err := godotenv.Load(); err != nil {
		return Config{}, err
	}

	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil

}
