package main

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/config"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/events"
	"activity-bot/internal/help"
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware"
	"activity-bot/internal/norm"
	"activity-bot/internal/predicate"
	"activity-bot/internal/rest"
	"activity-bot/internal/stats"
	"activity-bot/internal/summon"
	"context"
	"os"
	"os/signal"
	"time"

	fsm "github.com/fluffur/botapi-fsm"
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
	loc := translator.Default()
	_ = loc

	chatRepository := postgres.NewChatRepository(queries)
	userRepository := postgres.NewUserRepository(queries)
	chatMemberRepository := postgres.NewChatMemberRepository(queries)
	messageRepository := postgres.NewMessageRepository(queries)
	pmSessionRepository := postgres.NewPMSessionRepository(queries)
	permissionRepository := postgres.NewPermissionRepository(queries)
	normRepository := postgres.NewNormRepository(queries)
	statsRepository := postgres.NewStatsRepository(queries)
	restRepository := postgres.NewRestRepository(queries, pool)

	permissions := predicate.NewPermissionsChecker(permissionRepository, translator)
	rules := predicate.NewRuleChecker(chatMemberRepository, messageRepository)

	chatService := chat.NewService(chatRepository)
	chatMemberService := chatmember.NewService(chatRepository, userRepository, chatMemberRepository)
	statsService := stats.NewService(chatMemberRepository, normRepository, statsRepository)
	restService := rest.NewService(restRepository)

	registry := command.NewRegistry()

	bot.UseOuter(
		middleware.ChatMiddleware(chatRepository, pmSessionRepository),
		middleware.LocalizationMiddleware(translator),
		middleware.ChatMemberMiddleware(
			userRepository,
			chatMemberRepository,
			events.NewUsernameChangedNotifier(chatMemberRepository),
		),
		middleware.SaveMessageMiddleware(messageRepository),
	)

	bot.Use(
		botapi.Recover(),
		botapi.Timeout(time.Minute),
		botapi.Logging(),
	)

	summonStore := fsm.NewRedisJSONStore[summon.State, summon.StateData](client, "fsm:summon:", 5*time.Hour)
	summonFSM := fsm.New(
		summonStore,
		summon.StateIdle,
		fsm.WithKeyFunc[summon.State, summon.StateData](fsm.ChatSenderKey),
		fsm.WithUpdateKeyFunc[summon.State, summon.StateData](fsm.ChatSenderUpdateKey),
	)

	help.NewHandler(bot, permissions, registry, cfg.CommandsURL, cfg.DeveloperUsername).Register(registry)
	summon.NewHandler(bot, permissions, chatService, chatMemberService, summonFSM).Register(registry)
	norm.NewHandler(bot, permissions, rules, normRepository).Register(registry)
	stats.NewHandler(bot, permissions, rules, statsService).Register(registry)
	rest.NewHandler(bot, permissions, rules, restService, chatMemberService).Register(registry)
	events.NewHandler(bot, translator, chatMemberService).Register()

	log.Info("Starting bot")

	if err := bot.Run(ctx); err != nil {
		log.Fatal("Run", zap.Error(err))
	}
}
