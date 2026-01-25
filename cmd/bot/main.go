package main

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/call"
	"activity-bot/internal/chat"
	"activity-bot/internal/chat/member"
	"activity-bot/internal/command"
	"activity-bot/internal/config"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/exempt"
	"activity-bot/internal/filters"
	"activity-bot/internal/help"
	"activity-bot/internal/logger"
	msg "activity-bot/internal/message"
	"activity-bot/internal/stats"
	"activity-bot/internal/user"
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/chatmember"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		panic("Config load failed: " + err.Error())
	}

	logger.Init(cfg.Debug)

	b, err := gotgbot.NewBot(cfg.BotToken, &gotgbot.BotOpts{})
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	_, err = b.SetMyCommands([]gotgbot.BotCommand{
		{Command: "stats", Description: "📈 Отчёт за неделю"},
		{Command: "norm", Description: "📊 Норма сообщений"},
		{Command: "rest", Description: "💤 Статус реста"},
		{Command: "role", Description: "🎭 Узнать свою роль"},
		{Command: "roles", Description: "📜 Список всех ролей"},
		{Command: "admins", Description: "👮 Админы бота"},
		{Command: "update", Description: "🔄 Обновить данные чата"},
		{Command: "newbie", Description: "🐣 Срок новичка"},
		{Command: "help", Description: "❓ Помощь"},
	}, nil)
	if err != nil {
		slog.Warn("failed to set bot commands", "error", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		panic("failed to create pool: " + err.Error())
	}

	queries := db.New(pool)

	dp := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			slog.Error("an error occurred while handling update", "error", err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dp, &ext.UpdaterOpts{})

	statsRepository := postgres.NewStatsRepository(queries)
	exemptRepository := postgres.NewExemptRepository(queries, pool)
	memberRepository := postgres.NewMemberRepository(queries)
	userRepository := postgres.NewUserRepository(queries)
	chatRepository := postgres.NewChatRepository(queries)
	adminRepository := postgres.NewAdminRepository(queries)
	messageRepository := postgres.NewMessageRepository(queries)

	statsService := stats.NewService(statsRepository)
	exemptService := exempt.NewService(exemptRepository)
	chatService := chat.NewService(chatRepository, cfg.DefaultWeeklyNorm)
	userService := user.NewService(userRepository)
	memberService := member.NewService(memberRepository, chatRepository, userRepository, cfg.DefaultWeeklyNorm)
	adminService := admin.NewService(adminRepository)
	messageService := msg.NewService(messageRepository)
	cb := command.NewBuilder(userService, adminService)

	helpHandler := help.NewHandler(cfg.BotOwnerID)
	statsHandler := stats.NewHandler(statsService, exemptService, memberService)
	chatHandler := chat.NewHandler(chatService, adminService)
	exemptHandler := exempt.NewHandler(exemptService, userService, adminService, exempt.NewDateParser())
	adminHandler := admin.NewHandler(adminService, userService)
	messageHandler := msg.NewHandler(messageService, memberService)
	memberHandler := member.NewHandler(memberService, userService, adminService)
	callHandler := call.NewHandler(adminService, memberService)

	dp.AddHandler(cb.New("start", helpHandler.Start))
	dp.AddHandler(cb.New("help", helpHandler.Help))

	dp.AddHandler(cb.New("stats", statsHandler.ShowStats).
		SetAliases("отчёт", "отчет").
		SetMaxArgs(1).
		OnlyGroups(),
	)

	dp.AddHandler(cb.New("norm", chatHandler.ShowNorm).
		SetAliases("норма", "quota").
		OnlyGroups(),
	)
	dp.AddHandler(cb.New("norm", chatHandler.SetNorm).
		SetAliases("норма", "quota").
		SetTriggers("/", ".", "!", "+").
		AllowArgs().
		RequireAdmin().
		OnlyGroups().
		SetMaxArgs(1),
	)

	dp.AddHandler(cb.New("newbie", chatHandler.ShowNewbieThreshold).
		SetAliases("новичок", "newbies", "новички", "нью", "ньюхи").
		SetTriggers("/", ".", "!", "+", "").
		OnlyGroups(),
	)

	dp.AddHandler(cb.New("newbie", chatHandler.SetNewbieThreshold).
		SetAliases("новичок", "newbies", "новички", "нью", "ньюхи").
		SetTriggers("/", ".", "!", "+", "").
		AllowArgs().
		RequireAdmin().
		OnlyGroups().
		SetMaxArgs(1),
	)

	dp.AddHandler(cb.New("exempt", exemptHandler.Show).
		SetAliases("рест", "rest", "рэст").
		FallbackToSender().
		OnlyGroups().
		SetTriggers("/", ".", "!", "+", ""),
	)
	dp.AddHandler(cb.New("exempt", exemptHandler.Set).
		SetAliases("рест", "rest", "рэст").
		FallbackToSender().
		SetTriggers("/", ".", "!", "+", "").
		AllowArgs().
		OnlyGroups().
		SetMaxArgs(1),
	)
	dp.AddHandler(cb.New("-exempt", exemptHandler.End).
		FallbackToSender().
		OnlyGroups().
		SetAliases("-рест", "-rest", "-рэст").
		SetTriggers("/", ".", "!", ""),
	)
	dp.AddHandler(handlers.NewCallback(callbackquery.Prefix("approve:"), exemptHandler.ApproveExemptRequest))
	dp.AddHandler(handlers.NewCallback(callbackquery.Prefix("reject:"), exemptHandler.RejectExemptRequest))

	dp.AddHandler(cb.New("admins", adminHandler.ListAdmins).
		SetAliases("админы", "админчики", "администраторы", "адмы", "модеры", "mods").OnlyGroups(),
	)

	dp.AddHandler(cb.New("администратор", adminHandler.AddAdmin).
		SetAliases("админ", "admin", "адм", "модер").
		SetTriggers("/", ".", "!", "+").
		OnlyGroups().
		RequireAdmin(),
	)

	dp.AddHandler(cb.New("-администратор", adminHandler.RemoveAdmin).
		SetAliases("-админ", "-admin", "-адм", "-модер", "-mod").
		SetTriggers("/", ".", "!", "+", "").
		OnlyGroups().
		RequireCreator(),
	)

	dp.AddHandler(cb.New("обновить чат", memberHandler.UpdateMembersList).
		OnlyGroups().
		RequireAdmin().
		SetAliases("update chat", "update"),
	)

	dp.AddHandler(cb.New("роль", memberHandler.ShowRole).
		OnlyGroups().
		SetAliases("role", "title"),
	)

	dp.AddHandler(cb.New("роль", memberHandler.SetRole).
		SetAliases("role", "title").
		SetTriggers("/", ".", "!", "+").
		OnlyGroups().
		AllowArgs().
		SetMaxArgs(1),
	)

	dp.AddHandler(cb.New("роли", memberHandler.ListRoles).
		SetAliases("roles", "titles").OnlyGroups(),
	)

	dp.AddHandler(cb.New("call", callHandler.Call).
		SetAliases("калл", "колл").
		AllowArgs().
		OnlyGroups().
		RequireAdmin().
		SetMaxArgs(1),
	)

	dp.AddHandler(handlers.NewMessage(filters.OnlyGroupsText, messageHandler.Message))
	dp.AddHandler(handlers.NewMessage(message.NewChatMembers, memberHandler.OnJoinMember))
	dp.AddHandler(handlers.NewMessage(message.LeftChatMember, memberHandler.OnLeftMember))
	dp.AddHandler(handlers.NewMyChatMember(chatmember.NewStatus("administrator"), memberHandler.OnBotPromote))

	if cfg.WebhookURL != "" {
		webhookOpts := ext.WebhookOpts{
			ListenAddr:  fmt.Sprintf("0.0.0.0:%d", cfg.HTTPPort),
			SecretToken: cfg.WebhookSecretToken,
		}

		err = updater.StartWebhook(b, cfg.WebhookPath, webhookOpts)
		if err != nil {
			panic("failed to start webhook: " + err.Error())
		}

		err = updater.SetAllBotWebhooks(cfg.WebhookURL, &gotgbot.SetWebhookOpts{
			MaxConnections:     100,
			DropPendingUpdates: true,
			SecretToken:        cfg.WebhookSecretToken,
		})
		if err != nil {
			panic("failed to set webhook: " + err.Error())
		}

		slog.Info("Bot has been started with webhooks", "bot_username", b.User.Username, "url", cfg.WebhookURL, "port", cfg.HTTPPort)
	} else {
		err = updater.StartPolling(b, &ext.PollingOpts{
			DropPendingUpdates: true,
			GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
				Timeout: 9,
				RequestOpts: &gotgbot.RequestOpts{
					Timeout: time.Second * 10,
				},
			},
		})
		if err != nil {
			panic("failed to start polling: " + err.Error())
		}
		slog.Info("Bot has been started with long polling", "bot_username", b.User.Username)
	}

	updater.Idle()
}
