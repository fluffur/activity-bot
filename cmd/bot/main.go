package main

import (
	"activity-bot/internal/config"
	db "activity-bot/internal/db/sqlc"
	"activity-bot/internal/events"
	"activity-bot/internal/help"
	"activity-bot/internal/i18n"
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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log, _ := zap.NewDevelopment()
	defer func() { _ = log.Sync() }()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Load config", zap.Error(err))
	}

	store, err := storage.Open(cfg.StoragePath)
	if err != nil {
		log.Fatal("Open storage", zap.Error(err))
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

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatal("Connect to database", zap.Error(err))
	}

	queries := db.New(pool)
	_ = queries

	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisADDR,
	})
	defer func() { _ = client.Close() }()

	translator, err := i18n.New()
	if err != nil {
		log.Fatal("Create translator", zap.Error(err))
	}

	bot.Use(botapi.Recover(), botapi.Timeout(time.Minute), botapi.Logging())

	help.NewHandler(bot, translator).Register()
	events.NewHandler(bot, translator, events.NewService()).Register()

	log.Info("Starting bot")
	if err := bot.Run(ctx); err != nil {
		log.Fatal("Run", zap.Error(err))
	}
}
