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
	"activity-bot/internal/rest"
	"activity-bot/internal/user"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/cohesion-org/deepseek-go"
	"github.com/glebarez/sqlite"
	"github.com/gotd/td/telegram"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

type App struct {
	Config      config.Config
	Pool        *pgxpool.Pool
	Rdb         *redis.Client
	Deepseek    *deepseek.Client
	Bot         *gotgbot.Bot
	Dp          *ext.Dispatcher
	Updater     *ext.Updater
	AsyncClient *asynq.Client
	AsyncServer *asynq.Server
	GotdClient  *telegram.Client
	GotdReady   chan struct{}

	MemberService *member.Service
	AdminService  *admin.Service
	UserService   *user.Service
	ChatService   *chat.Service
	RestService   *rest.Service
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

	gotdReady := make(chan struct{})
	newClient, err := gotgproto.NewClient(cfg.AppID, cfg.AppHash, gotgproto.ClientTypeBot(cfg.BotToken), &gotgproto.ClientOpts{
		Session: sessionMaker.SqlSession(sqlite.Open("activity_bot")),
	})
	if err != nil {
		return nil, err
	}
	gotdClient := telegram.NewClient(cfg.AppID, cfg.AppHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{Path: cfg.SessionPath},
		NoUpdates:      true, // Disable updates to save memory
	})

	queries := db.New(pool)

	memberRepo := postgres.NewMemberRepository(queries)
	adminRepo := postgres.NewAdminRepository(queries)
	chatRepo := postgres.NewChatRepository(queries)
	userRepo := postgres.NewUserRepository(queries)
	restRepo := postgres.NewRestRepository(queries, pool)

	statusProvider := adapter.NewTelegramMemberStatusProvider(b)
	moderator := adapter.NewTelegramModerator(b)
	adminsProvider := adapter.NewTelegramChatMembersProvider(gotdClient, gotdReady)
	memberTagAdapter := adapter.NewMemberTagAdapter(b, chatRepo)

	userService := user.NewService(userRepo)
	memberService := member.NewService(memberRepo, chatRepo, userRepo, adminsProvider, memberTagAdapter)
	adminService := admin.NewService(adminRepo, statusProvider, moderator, cfg.BotOwnerID)
	chatService := chat.NewService(chatRepo)
	restService := rest.NewService(restRepo)

	dp := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			args := []any{"error", err}
			if ctx.EffectiveChat != nil {
				args = append(args, "chat_id", ctx.EffectiveChat.Id)
			}
			if ctx.EffectiveUser != nil {
				args = append(args, "user_id", ctx.EffectiveUser.Id)
			}
			if ctx.Update != nil {
				args = append(args, "update_id", ctx.Update.UpdateId)
			}
			slog.Error("an error occurred while handling update", args...)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	return &App{
		Config:   cfg,
		Pool:     pool,
		Rdb:      rdb,
		Deepseek: deepseek.NewClient(cfg.DeepseekAPIKey),
		Bot:      b,
		Dp:       dp,
		Updater: ext.NewUpdater(dp, &ext.UpdaterOpts{
			Logger: logger.L,
		}),
		AsyncClient:   client,
		AsyncServer:   srv,
		GotdClient:    gotdClient,
		GotdReady:     gotdReady,
		MemberService: memberService,
		AdminService:  adminService,
		UserService:   userService,
		ChatService:   chatService,
		RestService:   restService,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if err := a.setupBot(); err != nil {
		slog.Warn("bot setup failed", "error", err)
	}

	a.RegisterHandlers()
	go func() {
		var readyOnce sync.Once
		for {
			if err := a.startGotd(ctx, &readyOnce); err != nil {
				slog.Error("gotd start failed, reconnecting in 10s", "error", err)
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
		}
	}()

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
		c, err := a.ChatService.GetChat(ctx, p.ChatID)
		if err != nil {
			return err
		}
		m, err := a.MemberService.GetChatMember(ctx, p.ChatID, p.UserID)
		if err != nil {
			return err
		}

		if !c.TagsEnabled {
			if _, err = a.Bot.PromoteChatMember(p.ChatID, p.UserID, &gotgbot.PromoteChatMemberOpts{
				CanManageChat:   true,
				CanPostMessages: true,
				CanEditMessages: true,
			}); err != nil {
				return err
			}
		}

		if _, err = a.Bot.SendMessage(p.ChatID, fmt.Sprintf("Срок мута для участника %s подошёл к концу", helpers.RoleEmojiLink(m)), &gotgbot.SendMessageOpts{
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
			ParseMode: gotgbot.ParseModeHTML,
		}); err != nil {
			return err
		}

		return nil
	})

	var limiter = rate.NewLimiter(rate.Limit(25), 5)

	mux.HandleFunc("broadcast:post", func(ctx context.Context, t *asynq.Task) error {
		if err := limiter.Wait(ctx); err != nil {
			return err
		}

		var p model.BroadcastPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		kb := &gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
				{
					{
						Text:  "Открыть канал",
						Url:   helpers.TelegramChannelLink("FloodCMNews"),
						Style: "primary",
					},
				},
				{
					{
						Text:         "Отключить рассылку",
						CallbackData: "unsubscribe",
						Style:        "danger",
					},
				},
			},
		}

		_, err := a.Bot.CopyMessage(
			p.ChatID,
			p.FromChatID,
			p.MessageID,
			&gotgbot.CopyMessageOpts{
				ReplyMarkup: kb,
			},
		)

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
			CanManageTags:       true,
		},
	})
	if err != nil {
		return err
	}

	_, err = a.Bot.SetMyCommands(a.Config.BotCommands, nil)

	return err
}

func (a *App) startGotd(ctx context.Context, readyOnce *sync.Once) error {
	return a.GotdClient.Run(ctx, func(ctx context.Context) error {
		status, err := a.GotdClient.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !status.Authorized {
			if _, err := a.GotdClient.Auth().Bot(ctx, a.Config.BotToken); err != nil {
				return err
			}
		}
		readyOnce.Do(func() {
			close(a.GotdReady)
		})
		slog.Info("Gotd client has been started")
		<-ctx.Done()
		return nil
	})
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
			AllowedUpdates: []string{
				"message",
				"channel_post",
				"edited_channel_post",
				"callback_query",
			},
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
