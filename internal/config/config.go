package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	BotToken           string `env:"BOT_TOKEN"`
	DBDSN              string `env:"DB_DSN"`
	DefaultWeeklyNorm  int32  `env:"DEFAULT_WEEKLY_NORM" envDefault:"100"`
	WebhookURL         string `env:"WEBHOOK_URL"`
	WebhookPath        string `env:"WEBHOOK_PATH" envDefault:"telegram/webhook"`
	WebhookSecretToken string `env:"WEBHOOK_SECRET_TOKEN"`
	HTTPPort           int    `env:"HTTP_PORT" envDefault:"8080"`
	Debug              bool   `env:"DEBUG" envDefault:"false"`
	BotOwnerID         int64  `env:"BOT_OWNER_ID"`
	DeepseekAPIKey     string `env:"DEEPSEEK_API_KEY"`
	RedisADDR          string `env:"REDIS_ADDR" envDefault:"redis:6379"`
}

func Load() (Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil

}
