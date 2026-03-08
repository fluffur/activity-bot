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
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/cohesion-org/deepseek-go"
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
	Dispatcher  *ext.Dispatcher
	Updater     *ext.Updater
	AsyncClient *asynq.Client
	AsyncServer *asynq.Server

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

	queries := db.New(pool)

	memberRepo := postgres.NewMemberRepository(queries)
	adminRepo := postgres.NewAdminRepository(queries)
	chatRepo := postgres.NewChatRepository(queries)
	userRepo := postgres.NewUserRepository(queries)
	restRepo := postgres.NewRestRepository(queries, pool)

	statusProvider := adapter.NewTelegramMemberStatusProvider(b)
	moderator := adapter.NewTelegramModerator(b)
	adminsProvider := adapter.NewTelegramChatAdminsProvider(b)
	memberTagAdapter := adapter.NewMemberTagAdapter(b, chatRepo)

	userService := user.NewService(userRepo)
	memberService := member.NewService(memberRepo, chatRepo, userRepo, adminsProvider, cfg.DefaultNormWarn, memberTagAdapter)
	adminService := admin.NewService(adminRepo, statusProvider, moderator, cfg.BotOwnerID)
	chatService := chat.NewService(chatRepo, cfg.DefaultNormWarn)
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
		Config:     cfg,
		Pool:       pool,
		Rdb:        rdb,
		Deepseek:   deepseek.NewClient(cfg.DeepseekAPIKey),
		Bot:        b,
		Dispatcher: dp,
		Updater: ext.NewUpdater(dp, &ext.UpdaterOpts{
			Logger: logger.L,
		}),
		AsyncClient:   client,
		AsyncServer:   srv,
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

	if err := a.SyncRests(ctx); err != nil {
		slog.Error("Failed to sync rests", "error", err)
	}

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
			CanManageChat:   true,
			CanPostMessages: true,
			CanEditMessages: true,
		})
		if err != nil {
			return err
		}

		if _, err = a.Bot.SetChatAdministratorCustomTitle(p.ChatID, p.UserID, title, nil); err != nil {
			return err
		}
		if _, err = a.Bot.SendMessage(p.ChatID, fmt.Sprintf("Срок мута для участника %s подошёл к концу", helpers.LinkWithContent(m.User, fmt.Sprintf("%s (%s)", m.User.FirstName, m.CustomTitle))), &gotgbot.SendMessageOpts{
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
			ParseMode: gotgbot.ParseModeHTML,
		}); err != nil {
			return err
		}

		return nil
	})

	mux.HandleFunc("rest:expire", func(ctx context.Context, t *asynq.Task) error {
		var p model.RestExpirePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return err
		}

		m, err := a.MemberService.GetChatMember(ctx, p.ChatID, p.UserID)
		if err != nil || m.RestUntil == nil {
			return nil // User might have been deleted or rest ended manually
		}

		if m.RestUntil.Before(time.Now().In(helpers.MoscowLocation).Add(-30 * time.Second)) {
			return nil
		}

		if m.RestUntil.After(time.Now().In(helpers.MoscowLocation).Add(30 * time.Second)) {
			// This is a stale task (a newer rest was scheduled)
			return nil
		}

		admins, err := a.AdminService.GetAdminsEnsured(ctx, p.ChatID, a.MemberService.SyncChatMembers)
		if err != nil {
			return err
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Рест у участника %s подошёл к концу", helpers.LinkWithContent(m.User, fmt.Sprintf("%s (%s)", m.User.FirstName, m.CustomTitle))))
		for _, mod := range admins {
			sb.WriteString(helpers.Mention(mod.ID, "​"))
		}

		if _, err = a.Bot.SendMessage(p.ChatID, sb.String(), &gotgbot.SendMessageOpts{
			LinkPreviewOptions: &gotgbot.LinkPreviewOptions{
				IsDisabled: true,
			},
			ParseMode: gotgbot.ParseModeHTML,
		}); err != nil {
			return err
		}

		return a.RestService.EndMemberRest(ctx, p.ChatID, p.UserID)
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
						Text:  "Открыть пост в канале",
						Url:   helpers.TelegramMessageLink(p.FromChatID, p.MessageID, "FloodCMNews"),
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

func (a *App) SyncRests(ctx context.Context) error {
	rests, err := a.RestService.GetAllActiveRests(ctx)
	if err != nil {
		return err
	}

	slog.Info("Syncing rests", "count", len(rests))

	for _, r := range rests {
		if r.RestUntil.Before(time.Now()) {
			continue
		}

		payload, _ := json.Marshal(model.RestExpirePayload{
			ChatID: r.ChatID,
			UserID: r.UserID,
		})
		task := asynq.NewTask("rest:expire", payload)
		taskID := fmt.Sprintf("rest:expire:%d:%d", r.ChatID, r.UserID)
		if _, err := a.AsyncClient.Enqueue(task, asynq.ProcessAt(r.RestUntil), asynq.TaskID(taskID)); err != nil {
			if !errors.Is(err, asynq.ErrTaskIDConflict) {
				slog.Error("Failed to enqueue rest sync task", "chat_id", r.ChatID, "user_id", r.UserID, "error", err)
			}
		}
	}

	return nil
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
