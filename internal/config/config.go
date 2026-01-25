package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	BotToken           string `env:"BOT_TOKEN"`
	DBDSN              string `env:"DB_DSN"`
	DefaultWeeklyNorm  int32  `env:"DEFAULT_WEEKLY_NORM" envDefault:"100"`
	WebhookURL         string `env:"WEBHOOK_URL"`
	WebhookSecretToken string `env:"WEBHOOK_SECRET_TOKEN"`
	HTTPPort           int    `env:"HTTP_PORT" envDefault:"8080"`
	Debug              bool   `env:"DEBUG" envDefault:"false"`
}

func Load() (Config, error) {
	cfg := Config{}

	_ = godotenv.Load()

	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil

}
