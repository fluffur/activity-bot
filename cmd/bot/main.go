package main

import (
	"activity-bot/internal/adapter"
	"activity-bot/internal/admin"
	adminH "activity-bot/internal/admin/handler"
	callH "activity-bot/internal/call/handler"
	"activity-bot/internal/chat"
	chatH "activity-bot/internal/chat/handler"
	"activity-bot/internal/command"
	"activity-bot/internal/config"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/exempt"
	exemptH "activity-bot/internal/exempt/handler"
	"activity-bot/internal/filter"
	"activity-bot/internal/guard"
	helpH "activity-bot/internal/help/handler"
	"activity-bot/internal/logger"
	"activity-bot/internal/member"
	memberH "activity-bot/internal/member/handler"
	msg "activity-bot/internal/message"
	messageH "activity-bot/internal/message/handler"
	"activity-bot/internal/stats"
	statsH "activity-bot/internal/stats/handler"
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

	if _, err := b.SetMyDefaultAdministratorRightsWithContext(ctx, &gotgbot.SetMyDefaultAdministratorRightsOpts{
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
	}); err != nil {
		slog.Warn("failed to set default administrator rights", "error", err)
	}

	_, err = b.SetMyCommands([]gotgbot.BotCommand{
		{Command: "stats", Description: "📈 Недельный отчёт"},
		{Command: "norm", Description: "📊 Норма сообщений"},
		{Command: "rest", Description: "💤 Управление рестом"},
		{Command: "role", Description: "🎭 Роль пользователя"},
		{Command: "roles", Description: "📜 Список ролей"},
		{Command: "admins", Description: "👮 Администраторы бота"},
		{Command: "update", Description: "🔄 Обновить данные чата"},
		{Command: "newbie", Description: "🐣 Порог новичка"},
		{Command: "call", Description: "📞 Вызов участников"},
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

	adminsProvider := adapter.NewTelegramChatAdminsProvider(b)
	statusProvider := adapter.NewTelegramMemberStatusProvider(b)
	memberService := member.NewService(memberRepository, chatRepository, userRepository, adminsProvider, cfg.DefaultWeeklyNorm)
	adminService := admin.NewService(adminRepository, statusProvider, cfg.BotOwnerID)
	messageService := msg.NewService(messageRepository)

	dateParser := exempt.NewDateParser()

	helpHandler := helpH.New(cfg.BotOwnerID)
	statsHandler := statsH.New(statsService, exemptService, memberService)
	chatHandler := chatH.New(chatService, adminService, dateParser)
	exemptHandler := exemptH.New(exemptService, userService, adminService, dateParser)
	adminHandler := adminH.New(adminService, userService, memberService)
	messageHandler := messageH.New(messageService, memberService)
	memberHandler := memberH.New(memberService, userService)
	callHandler := callH.New(memberService)

	adminGuard := guard.NewAdminGuard(adminService)
	creatorGuard := guard.NewCreatorGuard(adminService)
	groupGuard := guard.OnlyGroups()

	dp.AddHandler(command.New("start", helpHandler.Start, userService))
	dp.AddHandler(command.New("help", helpHandler.Help, userService))

	dp.AddHandler(command.New("stats", statsHandler.ShowStats, userService).
		SetAliases("отчёт", "отчет").
		SetTriggers("/", ".", "!", "").
		SetMaxArgs(1).
		WithGuards(groupGuard),
	)

	dp.AddHandler(command.New("norm", chatHandler.ShowNorm, userService).
		SetAliases("норма", "quota").
		WithGuards(groupGuard),
	)
	dp.AddHandler(command.New("norm", chatHandler.SetNorm, userService).
		SetAliases("норма", "quota").
		SetTriggers("/", ".", "!", "+").
		AllowArgs().
		WithGuards(groupGuard, adminGuard).
		SetMaxArgs(1),
	)

	dp.AddHandler(command.New("олды кроме", chatHandler.SetOnlyNewbies, userService).
		WithGuards(groupGuard, adminGuard),
	)

	dp.AddHandler(command.New("newbie", chatHandler.ShowNewbieThreshold, userService).
		SetAliases("новичок", "newbies", "новички", "нью", "ньюхи").
		SetTriggers("/", ".", "!", "+").
		WithGuards(groupGuard, adminGuard),
	)

	dp.AddHandler(command.New("newbie", chatHandler.SetNewbieThreshold, userService).
		SetAliases("новичок", "newbies", "новички", "нью", "ньюхи").
		SetTriggers("/", ".", "!", "+").
		AllowArgs().
		WithGuards(groupGuard, adminGuard).
		SetMaxArgs(1),
	)

	dp.AddHandler(command.New("exempt", exemptHandler.Show, userService).
		SetAliases("рест", "rest", "рэст").
		FallbackToSender().
		WithGuards(groupGuard, adminGuard).
		SetTriggers("/", ".", "!", "+", ""),
	)
	dp.AddHandler(command.New("exempt", exemptHandler.Set, userService).
		SetAliases("рест", "rest", "рэст").
		FallbackToSender().
		SetTriggers("/", ".", "!", "+", "").
		AllowArgs().
		WithGuards(groupGuard).
		SetMaxArgs(1),
	)
	dp.AddHandler(command.New("-exempt", exemptHandler.End, userService).
		FallbackToSender().
		WithGuards(groupGuard).
		SetAliases("-рест", "-rest", "-рэст").
		SetTriggers("/", ".", "!", ""),
	)
	dp.AddHandler(handlers.NewCallback(callbackquery.Prefix("approve:"), exemptHandler.ApproveExemptRequest))
	dp.AddHandler(handlers.NewCallback(callbackquery.Prefix("reject:"), exemptHandler.RejectExemptRequest))

	dp.AddHandler(command.New("admins", adminHandler.ListAdmins, userService).
		SetAliases("админы", "админчики", "администраторы", "адмы", "модеры", "mods").
		WithGuards(groupGuard),
	)

	dp.AddHandler(command.New("администратор", adminHandler.IsAdmin, userService).
		SetAliases("админ", "admin", "адм", "модер").
		WithGuards(groupGuard).
		FallbackToSender(),
	)

	dp.AddHandler(command.New("администратор", adminHandler.AddAdmin, userService).
		SetAliases("админ", "admin", "адм", "модер").
		SetTriggers("+", "!+", "/+", ".+").
		WithGuards(groupGuard, adminGuard),
	)

	dp.AddHandler(command.New("-администратор", adminHandler.RemoveAdmin, userService).
		SetAliases("-админ", "-admin", "-адм", "-модер", "-mod").
		SetTriggers("/", ".", "!", "").
		WithGuards(groupGuard, creatorGuard),
	)

	dp.AddHandler(command.New("обновить чат", memberHandler.UpdateMembersList, userService).
		WithGuards(groupGuard).
		SetAliases("update chat", "update"),
	)

	dp.AddHandler(command.New("роли", memberHandler.ListRoles, userService).
		SetAliases("roles", "titles").
		WithGuards(groupGuard),
	)
	dp.AddHandler(command.New("-роль", memberHandler.DeleteRole, userService).
		WithGuards(groupGuard).
		SetAliases("-role", "-title").
		SetTriggers("/", ".", "!", ""),
	)
	dp.AddHandler(command.New("роль", memberHandler.ShowRole, userService).
		WithGuards(groupGuard).
		FallbackToSender().
		SetTriggers("/", ".", "!", "").
		SetAliases("role", "title"),
	)

	dp.AddHandler(command.New("роль", memberHandler.SetRole, userService).
		SetAliases("role", "title").
		SetTriggers("/", ".", "!", "+").
		WithGuards(groupGuard).
		AllowArgs().
		FallbackToSender().
		SetMaxArgs(1),
	)

	dp.AddHandler(command.New("call", callHandler.Call, userService).
		SetAliases("калл", "колл").
		AllowArgs().
		WithGuards(groupGuard, adminGuard).
		SetMaxArgs(1),
	)

	dp.AddHandler(handlers.NewMessage(message.LeftChatMember, memberHandler.OnLeftMember))
	dp.AddHandler(handlers.NewMessage(message.NewChatMembers, memberHandler.OnJoinMember))
	dp.AddHandler(handlers.NewMyChatMember(chatmember.NewStatus("administrator"), memberHandler.OnBotPromote))
	dp.AddHandler(handlers.NewMessage(filter.OnlyGroups, messageHandler.Message))

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
