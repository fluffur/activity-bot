package bot

import (
	"activity-bot/internal/adapter"
	"activity-bot/internal/admin"
	"activity-bot/internal/chat"
	"activity-bot/internal/config"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/helpers"
	"activity-bot/internal/logger"
	"activity-bot/internal/member"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/cohesion-org/deepseek-go"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Config      config.Config
	Pool        *pgxpool.Pool
	Rdb         *redis.Client
	Deepseek    *deepseek.Client
	Bot         *gotgbot.Bot
	Dispatcher  *ext.Dispatcher
	Updater     *ext.Updater
	AsyncClient *asynq.Client
	AsyncServer *asynq.Server

	MemberService *member.Service
	AdminService  *admin.Service
	UserService   *user.Service
	ChatService   *chat.Service
}

func NewApp(cfg config.Config) (*App, error) {
	ctx := context.Background()

	logger.Init(cfg.Debug)

	b, err := gotgbot.NewBot(cfg.BotToken, &gotgbot.BotOpts{})
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisADDR})
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.RedisADDR})
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisADDR},
		asynq.Config{Concurrency: 10},
	)

	queries := db.New(pool)

	memberRepo := postgres.NewMemberRepository(queries)
	adminRepo := postgres.NewAdminRepository(queries)
	chatRepo := postgres.NewChatRepository(queries)
	userRepo := postgres.NewUserRepository(queries)

	statusProvider := adapter.NewTelegramMemberStatusProvider(b)
	moderator := adapter.NewTelegramModerator(b)
	adminsProvider := adapter.NewTelegramChatAdminsProvider(b)

	userService := user.NewService(userRepo)
	memberService := member.NewService(memberRepo, chatRepo, userRepo, adminsProvider, cfg.DefaultNormWarn)
	adminService := admin.NewService(adminRepo, statusProvider, moderator)
	chatService := chat.NewService(chatRepo, cfg.DefaultNormWarn)

	if err := adminService.EnsureInitialDeveloper(ctx, cfg.BotOwnerID); err != nil {
		return nil, fmt.Errorf("failed to ensure initial developer: %w", err)
	}

	dp := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			slog.Error("an error occurred while handling update", "error", err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})

	return &App{
		Config:        cfg,
		Pool:          pool,
		Rdb:           rdb,
		Deepseek:      deepseek.NewClient(cfg.DeepseekAPIKey),
		Bot:           b,
		Dispatcher:    dp,
		Updater:       ext.NewUpdater(dp, &ext.UpdaterOpts{}),
		AsyncClient:   client,
		AsyncServer:   srv,
		MemberService: memberService,
		AdminService:  adminService,
		UserService:   userService,
		ChatService:   chatService,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if err := a.setupBot(); err != nil {
		slog.Warn("bot setup failed", "error", err)
	}

	a.RegisterHandlers()

	mux := a.registerWorkerHandlers()
	go func() {
		if err := a.AsyncServer.Start(mux); err != nil {
			slog.Error("Asynq server error", "error", err)
		}
	}()

	if a.Config.WebhookURL != "" {
		if err := a.startWebhook(); err != nil {
			return err
		}
	} else {
		if err := a.startPolling(); err != nil {
			return err
		}
		<-ctx.Done()
	}

	slog.Info("Shutting down...")
	a.AsyncServer.Shutdown()
	return a.Updater.Stop()
}

func (a *App) registerWorkerHandlers() *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc("role:restore", func(ctx context.Context, t *asynq.Task) error {
		var p model.RestoreRolePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		m, err := a.MemberService.GetChatMember(ctx, p.ChatID, p.UserID)
		title := m.CustomTitle
		if err != nil || title == "" {
			return err
		}
		_, err = a.Bot.PromoteChatMember(p.ChatID, p.UserID, &gotgbot.PromoteChatMemberOpts{
			CanPostMessages: true,
			CanEditMessages: true,
		})
		if err != nil {
			return err
		}

		if _, err = a.Bot.SetChatAdministratorCustomTitle(p.ChatID, p.UserID, title, nil); err != nil {
			return err
		}
		_, err = a.Bot.SendMessage(p.ChatID, fmt.Sprintf("Срок мута для участника %s подошёл к концу", helpers.LinkWithContent(m.User, fmt.Sprintf("%s (%s)", m.User.FirstName, m.CustomTitle))), &gotgbot.SendMessageOpts{
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
			ParseMode: gotgbot.ParseModeHTML,
		})

		return err
	})
	return mux
}

func (a *App) setupBot() error {
	_, err := a.Bot.SetMyDefaultAdministratorRights(&gotgbot.SetMyDefaultAdministratorRightsOpts{
		Rights: &gotgbot.ChatAdministratorRights{
			CanManageChat:       true,
			CanDeleteMessages:   true,
			CanManageVideoChats: true,
			CanRestrictMembers:  true,
			CanPromoteMembers:   true,
			CanChangeInfo:       true,
			CanInviteUsers:      true,
			CanPostStories:      true,
			CanEditStories:      true,
			CanDeleteStories:    true,
			CanPostMessages:     true,
			CanEditMessages:     true,
			CanPinMessages:      true,
			CanManageTopics:     true,
		},
	})
	if err != nil {
		return err
	}

	_, err = a.Bot.SetMyCommands(a.Config.BotCommands, nil)

	return err
}

func (a *App) startWebhook() error {
	webhookOpts := ext.WebhookOpts{
		ListenAddr:  fmt.Sprintf("0.0.0.0:%d", a.Config.HTTPPort),
		SecretToken: a.Config.WebhookSecretToken,
	}

	err := a.Updater.StartWebhook(a.Bot, a.Config.WebhookPath, webhookOpts)
	if err != nil {
		return err
	}

	err = a.Updater.SetAllBotWebhooks(a.Config.WebhookURL, &gotgbot.SetWebhookOpts{
		MaxConnections:     100,
		DropPendingUpdates: true,
		SecretToken:        a.Config.WebhookSecretToken,
	})
	if err != nil {
		return err
	}

	slog.Info("Bot has been started with webhooks", "bot_username", a.Bot.User.Username, "url", a.Config.WebhookURL)
	a.Updater.Idle()
	return nil
}

func (a *App) startPolling() error {
	err := a.Updater.StartPolling(a.Bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		return err
	}
	slog.Info("Bot has been started with long polling", "bot_username", a.Bot.User.Username)

	return nil
}

func (a *App) Close() {
	if a.Pool != nil {
		a.Pool.Close()
	}
	if a.Rdb != nil {
		_ = a.Rdb.Close()
	}
	if a.AsyncClient != nil {
		_ = a.AsyncClient.Close()
	}
}
