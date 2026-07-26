package main

import (
	"activity-bot/internal/application"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/config"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	fsm "github.com/fluffur/botapi-fsm"
	"github.com/gotd/log/logzap"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/gotd/botapi"
	"github.com/gotd/botapi/storage"
)

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
	defer func() { _ = store.Close() }()

	appFSM := fsm.NewRedisFSM[
		application.State,
		application.AppStateData,
	](
		redisClient,
		fmt.Sprintf("fsm:%s:app:", botKey),
		24*time.Hour,
		application.AppStateIdle,
		fsm.WithStrategy[
			application.State,
			application.AppStateData,
		](fsm.StrategySender),
	)

	rejectFSM := fsm.NewRedisFSM[
		application.RejectState,
		application.RejectStateData,
	](
		redisClient,
		fmt.Sprintf("fsm:%s:app:reject:", botKey),
		24*time.Hour,
		application.RejectStateIdle,
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

	handler := application.NewHandler(
		appFSM,
		rejectFSM,
		chatMemberService,
		application.NewRepository(redisClient),
		cfg.TargetChatID,
		cfg.ApplicationChatID,
		cfg.TargetChatLink,
		cfg.RolesPostLink,
	)

	bot.OnCommand(
		"start",
		"Запустить бота",
		handler.Start,
		botapi.ChatTypeIs(botapi.ChatTypePrivate),
	)

	bot.OnCallbackQuery(
		handler.StartCallback,
		botapi.CallbackData("app:new"),
	)

	bot.OnMessage(
		handler.ProcessRole,
		appFSM.State(application.AppStateAwaitRole),
		botapi.HasText(),
		botapi.ChatTypeIs(botapi.ChatTypePrivate),
	)

	bot.OnCallbackQuery(
		handler.ConfirmRules,
		botapi.CallbackData("app:confirm_rules"),
		appFSM.State(application.AppStateConfirmRole),
	)

	bot.OnCallbackQuery(
		handler.Accept,
		botapi.CallbackPrefix("app:accept:"),
	)

	bot.OnCallbackQuery(
		handler.Reject,
		botapi.CallbackPrefix("app:reject:"),
	)

	bot.OnMessage(
		handler.RejectMessage,
		rejectFSM.State(application.RejectStateAwaitRejectMessage),
		botapi.HasText(),
		botapi.Not(botapi.ChatTypeIs(botapi.ChatTypePrivate)),
	)

	botLog.Info("Starting application bot listener")

	return bot.Run(ctx)
}
