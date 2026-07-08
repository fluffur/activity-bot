package main

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/config"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/events"
	"activity-bot/internal/handler"
	"activity-bot/internal/help"
	"activity-bot/internal/i18n"
	"activity-bot/internal/message"
	"activity-bot/internal/middleware"
	"activity-bot/internal/moderation"
	"activity-bot/internal/norm"
	"activity-bot/internal/pmsession"
	"activity-bot/internal/predicate"
	"activity-bot/internal/register"
	"activity-bot/internal/rest"
	"activity-bot/internal/stats"
	"activity-bot/internal/summon"
	"activity-bot/internal/user"
	"context"
	"os"
	"os/signal"
	"time"

	glog "github.com/gotd/log"

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

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatal("Connect to database", zap.Error(err))
	}

	queries := db.New(pool)

	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisADDR,
	})
	defer func() { _ = client.Close() }()

	chatRepository := postgres.NewChatRepository(queries)
	userRepository := postgres.NewUserRepository(queries)
	chatMemberRepository := postgres.NewChatMemberRepository(queries)
	messageRepository := postgres.NewMessageRepository(queries)
	pmSessionRepository := postgres.NewPMSessionRepository(queries)
	permissionRepository := postgres.NewPermissionRepository(queries)
	normRepository := postgres.NewNormRepository(queries)
	statsRepository := postgres.NewStatsRepository(queries)
	restRepository := postgres.NewRestRepository(queries, pool)
	moderationRepository := postgres.NewModerationRepository(queries)

	chatService := chat.NewService(chatRepository)
	chatMemberService := chatmember.NewService(chatRepository, userRepository, chatMemberRepository)
	statsService := stats.NewService(chatMemberRepository, normRepository, statsRepository)
	restService := rest.NewService(restRepository)
	moderationService := moderation.NewService(moderationRepository, chatMemberRepository, cfg.DeveloperID)

	translator, err := i18n.New()
	if err != nil {
		log.Fatal("Create translator", zap.Error(err))
	}
	loc := translator.Default()

	permissions := predicate.NewPermissionsChecker(permissionRepository)
	rules := predicate.NewRuleChecker(chatMemberRepository, messageRepository)

	summonFSM := fsm.New(
		fsm.NewRedisJSONStore[summon.State, summon.StateData](client, "fsm:summon:", 5*time.Hour),
		summon.StateIdle,
		fsm.WithKeyFunc[summon.State, summon.StateData](fsm.ChatSenderKey),
		fsm.WithUpdateKeyFunc[summon.State, summon.StateData](fsm.ChatSenderUpdateKey),
	)

	registry := command.NewRegistry()

	handlers := []handler.Handler{
		help.NewHandler(registry, permissionRepository),
		summon.NewHandler(chatService, chatMemberService, summonFSM),
		norm.NewHandler(normRepository),
		stats.NewHandler(statsService),
		rest.NewHandler(restService, chatMemberService),
		moderation.NewHandler(moderationService, chatMemberService),
	}

	for _, h := range handlers {
		registry.Add(h.Actions())
	}

	var bot *botapi.Bot
	bot, err = botapi.New(cfg.BotToken, botapi.Options{
		AppID:   cfg.AppID,
		AppHash: cfg.AppHash,
		Logger:  logzap.New(log),
		Storage: store,
		OnStart: func(ctx context.Context) {
			registerBotCommands(ctx, bot, registry, loc)
		},
		FloodWait:                  true,
		DisableCommandRegistration: true,
	})
	if err != nil {
		log.Fatal("Create bot", zap.Error(err))
	}

	registerMiddlewares(
		bot,
		translator,
		chatRepository,
		pmSessionRepository,
		userRepository,
		chatMemberRepository,
		messageRepository,
	)

	register.Attach(bot, registry, permissions, rules)
	events.NewHandler(bot, translator, chatMemberService).Attach()

	log.Info("Starting bot")

	if err := bot.Run(ctx); err != nil {
		log.Fatal("Run", zap.Error(err))
	}
}

func registerMiddlewares(
	bot *botapi.Bot,
	translator *i18n.Translator,
	chatRepository chat.Repository,
	pmSessionRepository pmsession.Repository,
	userRepository user.Repository,
	chatMemberRepository chatmember.Repository,
	messageRepository message.Repository,
) {
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
}

func registerBotCommands(ctx context.Context, bot *botapi.Bot, registry *command.Registry, loc *i18n.Localizer) {
	if err := bot.SetMyCommands(ctx,
		register.BotCommands(registry, loc, command.ScopePrivate, false),
		botapi.WithCommandScope(botapi.BotCommandScopeAllPrivateChats()),
	); err != nil {
		glog.For(bot.Logger()).Error(ctx, "Set private commands", glog.Error(err))
	}

	if err := bot.SetMyCommands(ctx,
		register.BotCommands(registry, loc, command.ScopeGroup, false),
		botapi.WithCommandScope(botapi.BotCommandScopeAllGroupChats()),
	); err != nil {
		glog.For(bot.Logger()).Error(ctx, "Set group commands", glog.Error(err))
	}

	if err := bot.SetMyCommands(ctx,
		register.BotCommands(registry, loc, command.ScopeGroup, true),
		botapi.WithCommandScope(botapi.BotCommandScopeAllChatAdministrators()),
	); err != nil {
		glog.For(bot.Logger()).Error(ctx, "Set admin commands", glog.Error(err))
	}
}
