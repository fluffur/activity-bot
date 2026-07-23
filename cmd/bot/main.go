package main

import (
	"activity-bot/internal/crocodile"
	redis2 "activity-bot/internal/db/redis"
	"activity-bot/internal/rp"
	"activity-bot/internal/utils/tghtml"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
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
	AppStatePending   ApplicationState = "pending"
)

type RejectState string
type RejectStateData struct {
	UserID int64 `json:"user_id"`
}

const (
	RejectStateIdle               RejectState = ""
	RejectStateAwaitRejectMessage RejectState = "await_reject_message"
)

type AppStateData struct{}

type Application struct {
	UserID        int64  `json:"user_id"`
	Role          string `json:"role"`
	Username      string `json:"username"`
	ApplicationID int    `json:"application_id"`
}

func applicationKey(userID int64) string {
	return fmt.Sprintf("application:%d", userID)
}

func saveApplication(
	ctx context.Context,
	redisClient *redis.Client,
	app Application,
) error {
	data, err := json.Marshal(app)
	if err != nil {
		return err
	}

	return redisClient.Set(
		ctx,
		applicationKey(app.UserID),
		data,
		24*time.Hour,
	).Err()
}

func getApplication(
	ctx context.Context,
	redisClient *redis.Client,
	userID int64,
) (*Application, error) {
	data, err := redisClient.Get(
		ctx,
		applicationKey(userID),
	).Bytes()

	if err != nil {
		return nil, err
	}

	var app Application

	if err := json.Unmarshal(data, &app); err != nil {
		return nil, err
	}

	return &app, nil
}

func deleteApplication(
	ctx context.Context,
	redisClient *redis.Client,
	userID int64,
) error {
	return redisClient.Del(
		ctx,
		applicationKey(userID),
	).Err()
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

	appFSM := fsm.NewRedisFSM[
		ApplicationState,
		AppStateData,
	](
		redisClient,
		fmt.Sprintf("fsm:%s:app:", botKey),
		24*time.Hour,
		AppStateIdle,
		fsm.WithStrategy[
			ApplicationState,
			AppStateData,
		](fsm.StrategySender),
	)

	rejectFSM := fsm.NewRedisFSM[
		RejectState,
		RejectStateData,
	](
		redisClient,
		fmt.Sprintf("fsm:%s:app:reject:", botKey),
		24*time.Hour,
		RejectStateIdle,
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

	bot.OnCommand(
		"start",
		"Запустить бота",
		func(c *botapi.Context) error {
			sess, ok, err := appFSM.Get(c)
			if err != nil {
				return err
			}
			if ok && sess.State == AppStatePending {
				_, err := c.Reply("Пожалуйста, подождите пока вашу заявку обработают, перед тем как отправлять еще одну")
				return err
			}

			if err := appFSM.Enter(
				c,
				AppStateAwaitRole,
				AppStateData{},
			); err != nil {
				return err
			}

			_, err = c.Reply(
				tghtml.PatPatEmoji()+
					" Здравствуйте! Отправьте этому боту в сообщения желаемую роль\n\n"+tghtml.Link(cfg.RolesPostLink, "Посмотреть ролей флуда"),
				botapi.WithParseMode(botapi.ParseModeHTML),
				botapi.DisableWebPagePreview(),
			)

			return err
		},
		botapi.ChatTypeIs(botapi.ChatTypePrivate),
	)

	bot.OnMessage(
		func(c *botapi.Context) error {
			msg := c.Message()
			if msg == nil {
				return nil
			}

			role := predicate.NormalizeTag(msg.Text)

			if role == "" {
				_, err := c.Reply(
					"Пожалуйста, укажите корректную роль.",
				)
				return err
			}

			members, err := chatMemberService.ListHumanPresentChatMembers(
				c.Background(),
				cfg.TargetChatID,
			)

			if err != nil {
				log.Error(
					"Failed to list chat members",
					zap.Error(err),
				)

				_, err := c.Reply(
					"Произошла ошибка при проверке роли.",
				)

				return err
			}

			for _, m := range members {
				if strings.EqualFold(predicate.NormalizeTag(m.Tag), role) {
					_, err := c.Reply(
						"Эта роль уже занята. Пожалуйста, выберите другую.",
					)

					return err
				}
			}

			var senderID int64
			var username, firstname, lastname string
			if sender := c.Sender(); sender != nil {
				username = sender.Username
				firstname = sender.FirstName
				lastname = sender.LastName
				senderID = sender.ID
			} else {
				username = msg.Chat.Username
				firstname = msg.Chat.FirstName
				lastname = msg.Chat.LastName
				senderID = msg.Chat.ID
			}

			var userRef string

			if username != "" {
				userRef = "@" + username
			} else {
				userRef = strings.TrimSpace(
					firstname + " " + lastname,
				)

				if userRef == "" {
					userRef = fmt.Sprintf(
						"ID: %d",
						senderID,
					)
				}
			}

			app := Application{
				UserID:   senderID,
				Role:     role,
				Username: username,
			}

			appID := strconv.FormatInt(
				senderID,
				10,
			)

			adminMsg := fmt.Sprintf(
				"Новая заявка на вступление!\n\n"+
					"Роль: %s\n"+
					"Пользователь: %s",
				role,
				userRef,
			)

			sent, err := c.Bot.SendMessage(
				c,
				botapi.ID(cfg.ApplicationChatID),
				adminMsg,
				botapi.WithReplyMarkup(
					botapi.InlineKeyboard(
						botapi.InlineRow(
							botapi.InlineButtonData(
								"Принять",
								"app:accept:"+appID,
							),
							botapi.InlineButtonData(
								"Отклонить",
								"app:reject:"+appID,
							),
						),
					),
				),
			)
			if err != nil {
				return fmt.Errorf("send message: %w", err)
			}
			app.ApplicationID = sent.MessageID

			if err := appFSM.Enter(
				c,
				AppStatePending,
				AppStateData{},
			); err != nil {
				return err
			}

			if err := saveApplication(c, redisClient, app); err != nil {
				return err
			}

			_, err = c.Reply(
				"Ваша заявка отправлена на рассмотрение. Ожидайте ответа.",
			)

			return err
		},
		appFSM.State(AppStateAwaitRole),
		botapi.HasText(),
		botapi.ChatTypeIs(botapi.ChatTypePrivate),
	)

	bot.OnCallbackQuery(
		func(c *botapi.Context) error {
			cq := c.Update.CallbackQuery
			if cq == nil {
				return nil
			}

			prefix := "app:accept:"
			userID, err := strconv.ParseInt(
				strings.TrimPrefix(cq.Data, prefix),
				10,
				64,
			)

			if err != nil {
				return err
			}

			if err != nil {
				return err
			}

			application, err := getApplication(c, redisClient, userID)
			if err != nil {
				return err
			}

			_, err = c.Bot.SendMessage(
				c,
				botapi.ID(application.UserID),
				fmt.Sprintf(
					"Ваша заявка на роль %s была принята!\n%s",
					application.Role,
					cfg.TargetChatLink,
				),
			)

			if err != nil {
				log.Error(
					"notify applicant",
					zap.Error(err),
				)
			}

			if err := appFSM.ClearByKey(
				c.Background(),
				userID,
			); err != nil {
				return err
			}

			if err := deleteApplication(c, redisClient, application.UserID); err != nil {
				return err
			}
			_, _ = c.Bot.EditMessageText(
				c,
				botapi.ID(cq.Message.Chat.ID),
				cq.Message.MessageID,
				fmt.Sprintf(
					"Заявка на роль %s принята",
					application.Role,
				),
			)

			return c.AnswerCallback(
				botapi.WithCallbackText(
					"Заявка принята",
				),
			)
		},
		botapi.CallbackPrefix("app:accept:"),
	)

	bot.OnCallbackQuery(
		func(c *botapi.Context) error {
			cq := c.Update.CallbackQuery
			if cq == nil {
				return nil
			}

			userID, err := strconv.ParseInt(
				strings.TrimPrefix(cq.Data, "app:reject:"),
				10,
				64,
			)

			if err != nil {
				return err
			}

			sess, ok, err := appFSM.GetByKey(
				c.Background(),
				userID,
			)

			if err != nil {
				return err
			}

			if !ok || sess.State != AppStatePending {
				_ = c.AnswerCallback(
					botapi.WithCallbackText(
						"Заявка не найдена или устарела",
					),
				)

				return nil
			}

			if err := rejectFSM.Enter(
				c,
				RejectStateAwaitRejectMessage,
				RejectStateData{UserID: userID},
			); err != nil {
				return err
			}

			chatID, _ := c.Chat()
			_, err = c.Bot.SendMessage(
				c,
				chatID,
				"Введите причину отказа:",
			)

			return err
		},
		botapi.CallbackPrefix("app:reject:"),
	)

	bot.OnMessage(
		func(c *botapi.Context) error {
			msg := c.Message()
			if msg == nil {
				return nil
			}

			sess, ok, err := rejectFSM.Get(c)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}

			application, err := getApplication(c, redisClient, sess.Data.UserID)
			if err != nil {
				return err
			}

			reason := strings.TrimSpace(msg.Text)

			if reason == "" {
				_, err := c.Reply(
					"Причина не может быть пустой.",
				)

				return err
			}

			_, err = c.Bot.SendMessage(
				c,
				botapi.ID(application.UserID),
				fmt.Sprintf(
					"К сожалению, ваша заявка на роль %s была отклонена.\n\nПричина: %s",
					application.Role,
					reason,
				),
			)

			if err != nil {
				log.Error(
					"notify applicant rejection",
					zap.Error(err),
				)
			}

			if err := appFSM.ClearByKey(
				c.Background(),
				application.UserID,
			); err != nil {
				return err
			}

			if err := rejectFSM.Clear(c); err != nil {
				return err
			}

			_, _ = c.Bot.EditMessageText(
				c,
				botapi.ID(cfg.ApplicationChatID),
				application.ApplicationID,
				"Заявка отклонена\n\nПричина: "+reason,
			)

			_, err = c.Reply(
				"Заявка отклонена.",
			)

			return err
		},
		rejectFSM.State(RejectStateAwaitRejectMessage),
		botapi.HasText(),
		botapi.Not(botapi.ChatTypeIs(botapi.ChatTypePrivate)),
	)

	botLog.Info("Starting application bot listener")

	return bot.Run(ctx)
}
