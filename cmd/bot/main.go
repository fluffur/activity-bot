package main

import (
	"activity-bot/internal/config"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/events"
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/gotd/botapi"
	"github.com/gotd/botapi/storage"
	"github.com/gotd/log/logzap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	log, _ := zap.NewDevelopment()
	defer func() { _ = log.Sync() }()
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}
	store, err := storage.Open(cfg.StoragePath)
	if err != nil {
		log.Fatal("Failed to open storage", zap.Error(err))
	}
	defer func() { _ = store.Close() }()

	bot, err := botapi.New(cfg.BotToken, botapi.Options{
		AppID:     cfg.AppID,
		AppHash:   cfg.AppHash,
		Logger:    logzap.New(log),
		Storage:   store,
		FloodWait: true,
	})
	if err != nil {
		log.Fatal("Create bot", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	queries := db.New(pool)
	_ = queries

	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisADDR,
	})

	defer func() { _ = client.Close() }()

	bot.Use(botapi.Recover(), botapi.Timeout(time.Minute), botapi.Logging())

	events.NewHandler(bot, events.NewService()).Register()

	log.Info("Starting bot")
	if err := bot.Run(ctx); err != nil {
		log.Fatal("Run", zap.Error(err))
	}
}
