package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"activity-bot/internal/ai"
	"activity-bot/internal/chat"
	chatHandler "activity-bot/internal/chat/handler"
	"activity-bot/internal/chatmember"
	chatMemberHandler "activity-bot/internal/chatmember/handler"
	"activity-bot/internal/command"
	"activity-bot/internal/config"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/events"
	"activity-bot/internal/handler"
	"activity-bot/internal/help"
	"activity-bot/internal/i18n"
	"activity-bot/internal/manage"
	"activity-bot/internal/marriage"
	"activity-bot/internal/message"
	"activity-bot/internal/middleware"
	"activity-bot/internal/moderation"
	"activity-bot/internal/norm"
	permissionHandler "activity-bot/internal/permission/handler"
	"activity-bot/internal/pmsession"
	"activity-bot/internal/predicate"
	"activity-bot/internal/register"
	"activity-bot/internal/rest"
	rpHandler "activity-bot/internal/rp/handler"
	"activity-bot/internal/stats"
	"activity-bot/internal/summon"
	"activity-bot/internal/user"
	userHandler "activity-bot/internal/user/handler"

	"github.com/cohesion-org/deepseek-go"
	fsm "github.com/fluffur/botapi-fsm"
	glog "github.com/gotd/log"
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

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatal("Connect to database", zap.Error(err))
	}
	defer pool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisADDR,
	})
	defer func() { _ = redisClient.Close() }()

	deepseekClient := deepseek.NewClient(cfg.DeepseekAPIKey)

	translator, err := i18n.New()
	if err != nil {
		log.Fatal("Create translator", zap.Error(err))
	}

	var wg sync.WaitGroup

	for _, token := range cfg.BotTokens {
		wg.Add(1)
		go func(tkn string) {
			defer wg.Done()

			if err := runBotInstance(ctx, log, cfg, pool, redisClient, deepseekClient, translator, tkn); err != nil {
				log.Error("Bot instance execution failed",
					zap.String("bot_prefix", tkn[:8]),
					zap.Error(err),
				)
			}
		}(token)
	}

	log.Info("All bot instances successfully initialized and running")
	wg.Wait()
	log.Info("All bot instances gracefully stopped")
}

func runBotInstance(
	ctx context.Context,
	log *zap.Logger,
	cfg config.Config,
	pool *pgxpool.Pool,
	redisClient *redis.Client,
	deepseekClient *deepseek.Client,
	translator *i18n.Translator,
	token string,
) error {
	botKey := token[:8]
	botLog := log.With(zap.String("bot_key", botKey))

	sessionsDir := filepath.Join(cfg.StoragePath, "sessions")
	if err := os.MkdirAll(sessionsDir, 0750); err != nil {
		return fmt.Errorf("failed to create sessions directory for %s: %w", botKey, err)
	}

	instanceStoragePath := filepath.Join(sessionsDir, fmt.Sprintf("session_%s.bbolt", botKey))
	store, err := storage.Open(instanceStoragePath)
	if err != nil {
		return fmt.Errorf("open storage for %s: %w", botKey, err)
	}
	defer func() { _ = store.Close() }()

	queries := db.New(pool)

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
	marriageRepository := postgres.NewMarriageRepository(queries)
	rpRepository := postgres.NewRPRepository(queries)

	chatService := chat.NewService(chatRepository, cfg.DeveloperID)
	chatMemberService := chatmember.NewService(chatRepository, userRepository, chatMemberRepository)
	statsService := stats.NewService(chatMemberRepository, normRepository, statsRepository)
	restService := rest.NewService(restRepository)
	marriageService := marriage.NewService(marriageRepository)
	moderationService := moderation.NewService(moderationRepository, chatMemberRepository, cfg.DeveloperID)

	loc := translator.Default()

	permissions := predicate.NewPermissionsChecker(permissionRepository, cfg.DeveloperID)
	rules := predicate.NewRuleChecker(chatMemberRepository, messageRepository)

	summonFSM := fsm.NewRedisFSM[summon.State, summon.StateData](
		redisClient,
		fmt.Sprintf("fsm:%s:summon:", botKey),
		5*time.Hour,
		summon.StateIdle,
	)
	statsFSM := fsm.NewRedisFSM[stats.State, stats.StateData](
		redisClient,
		fmt.Sprintf("fsm:%s:stats:", botKey),
		10*time.Hour,
		stats.StateIdle,
	)

	registry := command.NewRegistry()
	summonH := summon.NewHandler(chatService, chatMemberService, summonFSM)
	handlers := []handler.Handler{
		help.NewHandler(registry, permissionRepository),
		summonH,
		norm.NewHandler(normRepository),
		stats.NewHandler(statsService, normRepository, summonH, statsFSM),
		rest.NewHandler(restService, chatMemberService),
		moderation.NewHandler(moderationService, chatMemberService),
		permissionHandler.NewHandler(registry, permissionRepository),
		manage.NewHandler(chatService, pmSessionRepository),
		userHandler.NewHandler(userRepository),
		chatMemberHandler.NewHandler(chatMemberRepository, chatMemberService),
		marriage.NewHandler(marriageService, chatMemberService),
		chatHandler.NewHandler(chatService),
		ai.NewHandler(deepseekClient),
		rpHandler.NewHandler(rpRepository),
	}

	for _, h := range handlers {
		registry.Add(h.Actions())
	}

	var bot *botapi.Bot
	bot, err = botapi.New(token, botapi.Options{
		AppID:   cfg.AppID,
		AppHash: cfg.AppHash,
		Logger:  logzap.New(botLog),
		Storage: store,
		OnStart: func(ctx context.Context) {
			registerBotCommands(ctx, bot, registry, loc)
			registerDefaultAdminRights(ctx, bot)
		},
		FloodWait:                  true,
		DisableCommandRegistration: true,
	})
	if err != nil {
		return fmt.Errorf("create bot for %s: %w", botKey, err)
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

	register.Attach(bot, registry, permissions, rules, rpRepository)
	events.NewHandler(bot, translator, chatMemberService).Attach()

	botLog.Info("Starting bot instance listener...")
	return bot.Run(ctx)
}

func registerDefaultAdminRights(ctx context.Context, bot *botapi.Bot) {
	err := bot.SetMyDefaultAdministratorRights(ctx, botapi.ChatAdminRights{
		IsAnonymous:         false,
		CanManageChat:       true,
		CanDeleteMessages:   true,
		CanManageVideoChats: true,
		CanRestrictMembers:  true,
		CanPromoteMembers:   true,
		CanChangeInfo:       true,
		CanInviteUsers:      true,
		CanPostMessages:     true,
		CanEditMessages:     true,
		CanPinMessages:      true,
		CanManageTopics:     true,
	}, false)

	if err != nil {
		if strings.Contains(err.Error(), "RIGHTS_NOT_MODIFIED") {
			glog.For(bot.Logger()).Info(ctx, "Default admin rights are already up to date")
			return
		}

		glog.For(bot.Logger()).Error(ctx, "failed to register default admin rights", glog.Error(err))
		return
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
