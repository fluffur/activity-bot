package config

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/caarlos0/env/v11"
)

type Config struct {
	BotToken           string `env:"BOT_TOKEN"`
	DBDSN              string `env:"DB_DSN"`
	DefaultNormWarn    int32  `env:"DEFAULT_NORM_WARN" envDefault:"100"`
	WebhookURL         string `env:"WEBHOOK_URL"`
	WebhookPath        string `env:"WEBHOOK_PATH" envDefault:"telegram/webhook"`
	WebhookSecretToken string `env:"WEBHOOK_SECRET_TOKEN"`
	HTTPPort           int    `env:"HTTP_PORT" envDefault:"8080"`
	Debug              bool   `env:"DEBUG" envDefault:"false"`
	BotOwnerID         int64  `env:"BOT_OWNER_ID"`
	DeepseekAPIKey     string `env:"DEEPSEEK_API_KEY"`
	RedisADDR          string `env:"REDIS_ADDR" envDefault:"redis:6379"`
	BotCommands        []gotgbot.BotCommand
}

func Load() (Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	cfg.BotCommands = BotCommands

	return cfg, nil

}
