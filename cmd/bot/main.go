package main

import (
	"activity-bot/internal/crocodile"
	redis2 "activity-bot/internal/db/redis"
	"activity-bot/internal/rp"
	"context"
	"encoding/json"
	"errors"
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
	"github.com/davecgh/go-spew/spew"
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
	crocodileRepository := redis2.NewCrocodileRepository(redisClient)
	wordRepository := postgres.NewWordRepository(queries)

	chatService := chat.NewService(chatRepository, cfg.DeveloperID)
	chatMemberService := chatmember.NewService(chatRepository, userRepository, chatMemberRepository)
	statsService := stats.NewService(chatMemberRepository, normRepository, statsRepository)
	restService := rest.NewService(restRepository)
	marriageService := marriage.NewService(marriageRepository)
	moderationService := moderation.NewService(moderationRepository, chatMemberRepository, cfg.DeveloperID)
	crocodileService := crocodile.NewService(crocodileRepository, wordRepository)

	loc := translator.Default()

	permissions := predicate.NewPermissionsChecker(permissionRepository, cfg.DeveloperID)
	rules := predicate.NewRuleChecker(chatMemberRepository, messageRepository)

	var wg sync.WaitGroup

	for _, token := range cfg.BotTokens {
		wg.Add(1)

		go func(tkn string) {
			defer wg.Done()

			if err := runBotInstance(
				ctx,
				log,
				cfg,

				redisClient,
				deepseekClient,
				translator,

				tkn,

				chatRepository,
				userRepository,
				chatMemberRepository,
				messageRepository,
				pmSessionRepository,
				permissionRepository,
				normRepository,
				rpRepository,

				chatService,
				chatMemberService,
				statsService,
				restService,
				marriageService,
				moderationService,
				crocodileService,

				permissions,
				rules,

				loc,
			); err != nil {
				log.Error(
					"Bot instance execution failed",
					zap.String("bot_prefix", tkn[:8]),
					zap.Error(err),
				)
			}
		}(token)
	}

	if cfg.ApplicationBotToken != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := runApplicationBot(
				ctx,
				log,
				cfg,
				redisClient,
				chatMemberService,
			); err != nil {
				log.Error("Application bot failed", zap.Error(err))
			}
		}()
	}

	log.Info("All bot instances successfully initialized and running")
	wg.Wait()
	log.Info("All bot instances gracefully stopped")
}

func runBotInstance(
	ctx context.Context,
	log *zap.Logger,
	cfg config.Config,

	redisClient *redis.Client,
	deepseekClient *deepseek.Client,
	translator *i18n.Translator,

	token string,

	chatRepository chat.Repository,
	userRepository user.Repository,
	chatMemberRepository chatmember.Repository,
	messageRepository message.Repository,
	pmSessionRepository pmsession.Repository,
	permissionRepository *postgres.PermissionRepository,
	normRepository norm.Repository,
	rpRepository rp.Repository,

	chatService *chat.Service,
	chatMemberService *chatmember.Service,
	statsService *stats.Service,
	restService *rest.Service,
	marriageService *marriage.Service,
	moderationService *moderation.Service,
	crocodileService *crocodile.Service,

	permissions *predicate.PermissionChecker,
	rules *predicate.RuleChecker,

	loc *i18n.Localizer,
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
	rpFsm := fsm.NewRedisFSM[rp.State, rp.StateData](
		redisClient,
		fmt.Sprintf("fsm:%s:rp", botKey),
		10*time.Hour,
		rp.StateIdle,
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
		marriage.NewHandler(marriageService, chatService, chatMemberService),
		chatHandler.NewHandler(chatService),
		ai.NewHandler(deepseekClient),
		rpHandler.NewHandler(rpRepository, chatMemberService, userRepository, rpFsm),
		crocodile.NewHandler(crocodileService),
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
		summonH,
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
	summonH *summon.Handler,
) {
	bot.UseOuter(
		middleware.ChatMiddleware(chatRepository, pmSessionRepository),
		middleware.LocalizationMiddleware(translator),
		middleware.ChatMemberMiddleware(
			userRepository,
			chatMemberRepository,
			events.NewUsernameChangedNotifier(chatMemberRepository, summonH),
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

type ApplicationState string

const (
	AppStateIdle      ApplicationState = ""
	AppStateAwaitRole ApplicationState = "await_role"
)

type ApplicationStateData struct {
	Test bool
}

type PendingApplication struct {
	UserID   int64  `json:"user_id"`
	ChatID   int64  `json:"chat_id"`
	Role     string `json:"role"`
	Username string `json:"username"`
}

func runApplicationBot(
	ctx context.Context,
	log *zap.Logger,
	cfg config.Config,
	redisClient *redis.Client,
	chatMemberService *chatmember.Service,
) error {
	botKey := cfg.ApplicationBotToken[:8]
	botLog := log.With(
		zap.String("bot_key", botKey),
	)

	sessionsDir := filepath.Join(cfg.StoragePath, "sessions")
	if err := os.MkdirAll(sessionsDir, 0750); err != nil {
		return fmt.Errorf("create sessions dir: %w", err)
	}

	storePath := filepath.Join(
		sessionsDir,
		fmt.Sprintf("application_%s.bbolt", botKey),
	)

	store, err := storage.Open(storePath)
	if err != nil {
		return fmt.Errorf("open storage: %w", err)
	}
	defer store.Close()

	appFSM := fsm.NewRedisFSM[ApplicationState, ApplicationStateData](
		redisClient,
		fmt.Sprintf("fsm:%s:app:", botKey),
		24*time.Hour,
		AppStateIdle,
		fsm.WithStrategy[ApplicationState, ApplicationStateData](fsm.StrategySender),
	)

	var bot *botapi.Bot

	bot, err = botapi.New(cfg.ApplicationBotToken, botapi.Options{
		AppID:   cfg.AppID,
		AppHash: cfg.AppHash,
		Logger:  logzap.New(botLog),
		Storage: store,
		OnStart: func(ctx context.Context) {
			botLog.Info("Application bot started")
		},
		FloodWait: true,
	})
	if err != nil {
		return fmt.Errorf("create application bot: %w", err)
	}

	bot.OnMessage(func(c *botapi.Context) error {
		msg := c.Message()
		if msg == nil {
			return nil
		}

		if msg.Chat.Type != botapi.ChatTypePrivate {
			return nil
		}

		if strings.HasPrefix(msg.Text, "/start") {
			key, ok := fsm.SenderKey(c)
			fmt.Println("ENTER", key, ok)
			if err := appFSM.Enter(c, AppStateAwaitRole, ApplicationStateData{Test: true}); err != nil {
				return err
			}
			_, err := c.Reply("Здравствуйте, укажите желаемую роль")
			return err
		}
		key, ok := fsm.SenderKey(c)
		fmt.Println("GET", key, ok)
		session, ok, err := appFSM.Get(c)
		if err != nil {
			return err
		}
		spew.Dump(session, ok, err)
		if ok && session.State == AppStateAwaitRole {
			role := strings.TrimSpace(msg.Text)
			if role == "" {
				_, err := c.Reply("Пожалуйста, укажите корректную роль.")
				return err
			}

			members, err := chatMemberService.ListHumanPresentChatMembers(c.Background(), cfg.TargetChatID)
			if err != nil {
				log.Error("Failed to list chat members", zap.Error(err))
				_, err := c.Reply("Произошла ошибка при проверке роли. Пожалуйста, попробуйте позже.")
				return err
			}

			isOccupied := false
			for _, m := range members {
				if strings.EqualFold(m.Tag, role) {
					isOccupied = true
					break
				}
			}

			if isOccupied {
				_, err := c.Reply("Эта роль уже занята. Пожалуйста, выберите другую.")
				return err
			}

			if err := appFSM.Clear(c); err != nil {
				return err
			}

			sender := c.Sender()
			var username string
			var userRef string
			if sender != nil {
				username = sender.Username
				if username != "" {
					userRef = "@" + username
				} else {
					userRef = strings.TrimSpace(sender.FirstName + " " + sender.LastName)
					if userRef == "" {
						userRef = fmt.Sprintf("ID: %d", sender.ID)
					}
				}
			}

			appID := fmt.Sprintf("%d_%d", sender.ID, time.Now().UnixNano())

			pending := PendingApplication{
				UserID:   sender.ID,
				ChatID:   c.Update.Message.Chat.ID,
				Role:     role,
				Username: username,
			}

			pendingData, err := json.Marshal(pending)
			if err != nil {
				return err
			}

			redisKey := fmt.Sprintf("app:pending:%s", appID)
			if err := redisClient.Set(c.Background(), redisKey, pendingData, 7*24*time.Hour).Err(); err != nil {
				log.Error("Failed to save application to redis", zap.Error(err))
				_, err := c.Reply("Произошла ошибка при сохранении заявки. Пожалуйста, попробуйте позже.")
				return err
			}

			adminMsgText := fmt.Sprintf("Новая заявка на вступление!\nРоль: %s\nПользователь: %s", role, userRef)
			_, err = c.Bot.SendMessage(
				c,
				botapi.ID(cfg.ApplicationChatID),
				adminMsgText,
				botapi.WithReplyMarkup(
					botapi.InlineKeyboard(
						botapi.InlineRow(
							botapi.InlineButtonData("Принять", "app:accept:"+appID),
							botapi.InlineButtonData("Отклонить", "app:reject:"+appID),
						),
					),
				),
			)
			if err != nil {
				log.Error("Failed to send admin notification", zap.Error(err))
				_, err := c.Reply("Произошла ошибка при отправке заявки админам. Пожалуйста, попробуйте позже.")
				return err
			}

			_, err = c.Reply("Ваша заявка отправлена на рассмотрение. Ожидайте ответа.")
			return err
		}

		return nil
	})

	bot.OnCallbackQuery(func(c *botapi.Context) error {
		cq := c.Update.CallbackQuery
		if cq == nil {
			return nil
		}

		if strings.HasPrefix(cq.Data, "app:accept:") {
			appID := strings.TrimPrefix(cq.Data, "app:accept:")
			redisKey := fmt.Sprintf("app:pending:%s", appID)
			data, err := redisClient.Get(c.Background(), redisKey).Bytes()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					_ = c.AnswerCallback(botapi.WithCallbackText("Заявка не найдена или устарела."))
					return nil
				}
				return err
			}

			var pending PendingApplication
			if err := json.Unmarshal(data, &pending); err != nil {
				return err
			}

			_ = redisClient.Del(c.Background(), redisKey)

			applicantMsg := fmt.Sprintf("Ваша заявка на роль %s была принята!\nСсылка на чат: %s", pending.Role, cfg.TargetChatLink)
			_, err = c.Bot.SendMessage(c, botapi.ID(pending.ChatID), applicantMsg)
			if err != nil {
				log.Error("Failed to notify applicant of acceptance", zap.Error(err))
			}

			var userRef string
			if pending.Username != "" {
				userRef = "@" + pending.Username
			} else {
				userRef = fmt.Sprintf("ID: %d", pending.UserID)
			}
			newAdminText := fmt.Sprintf("Заявка от %s на роль %s принята.", userRef, pending.Role)
			_, _ = c.Bot.EditMessageText(
				c,
				botapi.ID(cq.Message.Chat.ID),
				cq.Message.MessageID,
				newAdminText,
			)

			_ = c.AnswerCallback(botapi.WithCallbackText("Заявка принята!"))
			return nil
		}

		if strings.HasPrefix(cq.Data, "app:reject:") {
			appID := strings.TrimPrefix(cq.Data, "app:reject:")
			redisKey := fmt.Sprintf("app:pending:%s", appID)
			data, err := redisClient.Get(c.Background(), redisKey).Bytes()
			if err != nil {
				if errors.Is(err, redis.Nil) {
					_ = c.AnswerCallback(botapi.WithCallbackText("Заявка не найдена или устарела."))
					return nil
				}
				return err
			}

			var pending PendingApplication
			if err := json.Unmarshal(data, &pending); err != nil {
				return err
			}

			_ = redisClient.Del(c.Background(), redisKey)

			applicantMsg := fmt.Sprintf("К сожалению, ваша заявка на роль %s была отклонена.", pending.Role)
			_, err = c.Bot.SendMessage(c, botapi.ID(pending.ChatID), applicantMsg)
			if err != nil {
				log.Error("Failed to notify applicant of rejection", zap.Error(err))
			}

			var userRef string
			if pending.Username != "" {
				userRef = "@" + pending.Username
			} else {
				userRef = fmt.Sprintf("ID: %d", pending.UserID)
			}
			newAdminText := fmt.Sprintf("Заявка от %s на роль %s отклонена.", userRef, pending.Role)
			_, _ = c.Bot.EditMessageText(
				c,
				botapi.ID(cq.Message.Chat.ID),
				cq.Message.MessageID,
				newAdminText,
			)

			_ = c.AnswerCallback(botapi.WithCallbackText("Заявка отклонена!"))
			return nil
		}

		return nil
	})

	botLog.Info("Starting application bot listener")

	return bot.Run(ctx)
}
