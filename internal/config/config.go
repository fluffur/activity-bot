package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	BotToken         string `env:"BOT_TOKEN"`
	DBDSN            string `env:"DB_DSN"`
	Debug            bool   `env:"DEBUG" envDefault:"false"`
	BotOwnerID       int64  `env:"BOT_OWNER_ID"`
	BotOwnerUsername string `env:"BOT_OWNER_USERNAME"`
	AppID            int    `env:"APP_ID"`
	AppHash          string `env:"APP_HASH"`
	StoragePath      string `env:"STORAGE_PATH" envDefault:"session.bbolt"`
	DeepseekAPIKey   string `env:"DEEPSEEK_API_KEY"`
	RedisADDR        string `env:"REDIS_ADDR" envDefault:"redis:6379"`
	UniquePrefix     string `env:"UNIQUE_PREFIX" envDefault:"фм"`
}

func Load() (Config, error) {
	cfg := Config{}

	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil

}
