package main

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/config"
	"activity-bot/internal/db/repository"
	db "activity-bot/internal/db/sqlc"
	"activity-bot/internal/events"
	"activity-bot/internal/help"
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware"
	"activity-bot/internal/predicate"
	"activity-bot/internal/summon"
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/gotd/log/logzap"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/gotd/botapi"
	"github.com/gotd/botapi/storage"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log, _ := zap.NewProduction()
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
		AppID:                      cfg.AppID,
		AppHash:                    cfg.AppHash,
		Logger:                     logzap.New(log),
		Storage:                    store,
		FloodWait:                  true,
		DisableCommandRegistration: true,
	})
	if err != nil {
		log.Fatal("Create bot", zap.Error(err))
	}

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatal("Connect to database", zap.Error(err))
	}

	queries := db.New(pool)

	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisADDR,
	})
	defer func() { _ = client.Close() }()

	translator, err := i18n.New()
	if err != nil {
		log.Fatal("Create translator", zap.Error(err))
	}

	chatRepository := repository.NewChatRepository(queries)
	userRepository := repository.NewUserRepository(queries)
	chatMemberRepository := repository.NewChatMemberRepository(queries)
	messageRepository := repository.NewMessageRepository(queries)
	pmSessionRepository := repository.NewPMSessionRepository(queries)
	permissionRepository := repository.NewPermissionRepository(queries)

	chatMemberService := chatmember.NewService(chatRepository, userRepository, chatMemberRepository)
	permissions := predicate.NewPermissionsChecker(permissionRepository, translator)

	bot.UseOuter(
		middleware.ChatMiddleware(chatRepository, pmSessionRepository),
		middleware.ChatMemberMiddleware(userRepository, chatMemberRepository),
	)

	bot.Use(
		botapi.Recover(),
		botapi.Timeout(time.Minute),
		botapi.Logging(),
	)

	help.NewHandler(bot, translator, permissions, cfg.CommandsURL, cfg.DeveloperUsername).Register()
	summon.NewHandler(bot, translator, permissions, chatMemberService).Register()

	events.NewHandler(bot, translator, messageRepository, chatMemberService).Register()

	log.Info("Starting bot")

	if err := bot.Run(ctx); err != nil {
		log.Fatal("Run", zap.Error(err))
	}
}
