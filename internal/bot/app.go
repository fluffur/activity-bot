package bot

import (
	"activity-bot/internal/config"
	"activity-bot/internal/logger"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/cohesion-org/deepseek-go"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Config     config.Config
	Pool       *pgxpool.Pool
	Rdb        *redis.Client
	Deepseek   *deepseek.Client
	Bot        *gotgbot.Bot
	Dispatcher *ext.Dispatcher
	Updater    *ext.Updater
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

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisADDR,
	})

	dsClient := deepseek.NewClient(cfg.DeepseekAPIKey)

	dp := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			slog.Error("an error occurred while handling update", "error", err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})

	updater := ext.NewUpdater(dp, &ext.UpdaterOpts{})

	return &App{
		Config:     cfg,
		Pool:       pool,
		Rdb:        rdb,
		Deepseek:   dsClient,
		Bot:        b,
		Dispatcher: dp,
		Updater:    updater,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if err := a.setupBot(); err != nil {
		slog.Warn("bot setup failed", "error", err)
	}

	a.RegisterHandlers()

	stopFunc := func() error { return nil }

	if a.Config.WebhookURL != "" {
		if err := a.startWebhook(); err != nil {
			return err
		}
		stopFunc = func() error {
			return a.Updater.Stop()
		}
	} else {
		if err := a.startPolling(); err != nil {
			return err
		}
		stopFunc = func() error {
			return a.Updater.Stop()
		}
	}

	<-ctx.Done()
	slog.Info("Shutting down bot...")

	if err := stopFunc(); err != nil {
		slog.Error("Failed to stop updater", "error", err)
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
	// Removed Idle() logic from here as it is handled in Run
	return nil
}

func (a *App) Close() {
	if a.Pool != nil {
		a.Pool.Close()
	}
	if a.Rdb != nil {
		_ = a.Rdb.Close()
	}
}
