package main

import (
	"activity-bot/internal/adapter"
	"activity-bot/internal/admin"
	adminH "activity-bot/internal/admin/handler"
	callH "activity-bot/internal/call/handler"
	"activity-bot/internal/chat"
	chatH "activity-bot/internal/chat/handler"
	"activity-bot/internal/cmd"
	"activity-bot/internal/config"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/filter"
	"activity-bot/internal/guard"
	helpH "activity-bot/internal/help/handler"
	"activity-bot/internal/logger"
	"activity-bot/internal/member"
	memberH "activity-bot/internal/member/handler"
	msg "activity-bot/internal/message"
	messageH "activity-bot/internal/message/handler"
	"activity-bot/internal/rest"
	restH "activity-bot/internal/rest/handler"
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
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/chatmember"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/genai"
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
		{Command: "ladder", Description: "🪜 Максимальная лесенка сообщений"},
		{Command: "rest", Description: "💤 Управление рестом"},
		{Command: "role", Description: "🎭 Роль пользователя"},
		{Command: "whoami", Description: "ℹ️ Информация о себе"},
		{Command: "whoareu", Description: "👤️ Информация об участнике"},
		{Command: "role", Description: "🎭 Роль пользователя"},
		{Command: "roles", Description: "📜 Список ролей"},
		{Command: "admins", Description: "👮 Администраторы бота"},
		{Command: "is_admin", Description: "👮 Проверить статус администратора"},
		{Command: "update", Description: "🔄 Обновить данные чата"},
		{Command: "newbie", Description: "🐣 Порог новичка"},
		{Command: "all", Description: "📞 Вызов участников"},
		{Command: "help", Description: "❓ Помощь"},
	}, nil)
	if err != nil {
		slog.Warn("failed to set bot commands", "error", err)
	}

	pool, err := pgxpool.New(ctx, cfg.DBDSN)
	if err != nil {
		panic("failed to create pool: " + err.Error())
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisADDR,
	})
	defer func(rdb *redis.Client) {
		err := rdb.Close()
		if err != nil {
			panic("Failed to close Redis connection: " + err.Error())
		}
	}(rdb)

	queries := db.New(pool)

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.GeminiAPIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		panic("failed to init gemini client " + err.Error())
	}

	dp := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			slog.Error("an error occurred while handling update", "error", err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})
	updater := ext.NewUpdater(dp, &ext.UpdaterOpts{})

	statsRepository := postgres.NewStatsRepository(queries)
	restRepository := postgres.NewRestRepository(queries, pool)
	memberRepository := postgres.NewMemberRepository(queries)
	userRepository := postgres.NewUserRepository(queries)
	chatRepository := postgres.NewChatRepository(queries)
	adminRepository := postgres.NewAdminRepository(queries)
	messageRepository := postgres.NewMessageRepository(queries)

	statsService := stats.NewService(statsRepository)
	restService := rest.NewService(restRepository)
	chatService := chat.NewService(chatRepository, cfg.DefaultWeeklyNorm)
	userService := user.NewService(userRepository)

	adminsProvider := adapter.NewTelegramChatAdminsProvider(b)
	statusProvider := adapter.NewTelegramMemberStatusProvider(b)
	memberService := member.NewService(memberRepository, chatRepository, userRepository, adminsProvider, cfg.DefaultWeeklyNorm)
	adminService := admin.NewService(adminRepository, statusProvider, cfg.BotOwnerID)
	messageService := msg.NewService(messageRepository)

	dateParser := rest.NewDateParser()

	helpHandler := helpH.New(cfg.BotOwnerID)
	statsHandler := statsH.New(statsService, restService, memberService)
	chatHandler := chatH.New(chatService, adminService, dateParser)
	restHandler := restH.New(restService, userService, adminService, dateParser)
	adminHandler := adminH.New(adminService, userService, memberService)
	messageHandler := messageH.New(messageService, memberService, chatService, client)
	memberHandler := memberH.New(memberService, userService)
	callHandler := callH.New(memberService)

	adminGuard := guard.NewAdminGuard(adminService)
	creatorGuard := guard.NewCreatorGuard(adminService)
	groupGuard := guard.OnlyGroups()
	rateLimiterGuard := guard.NewRateLimiter(rdb, 1, 10*time.Second)

	cf := cmd.NewFactory(userService, "/", "!", ".")

	dp.AddHandler(cf.New(helpHandler.Start, "start"))
	dp.AddHandler(cf.New(helpHandler.Help, "help"))

	dp.AddHandler(cf.New(statsHandler.ShowStats, "stats", "отчёт", "отчет").
		AddTriggers("").
		SetArgsCount(1).
		WithGuards(groupGuard, rateLimiterGuard),
	)
	dp.AddHandler(cf.New(statsHandler.WhoAmI, "whoami", "ктоя", "я кто").
		AddTriggers("").
		WithGuards(groupGuard),
	)
	dp.AddHandler(cf.New(statsHandler.WhoAreYou, "whoareu", "ктоты", "тыкто").
		AddTriggers("").
		WithGuards(groupGuard),
	)

	dp.AddHandler(cf.New(chatHandler.ShowNorm, "norm", "норма какая", "а норма какая", "норма", "норма?", "quota", "какая норма", "а какая норма").
		AddTriggers("").
		WithGuards(groupGuard),
	)
	dp.AddHandler(cf.New(chatHandler.SetNorm, "norm", "норма", "quota").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)

	dp.AddHandler(cf.New(memberHandler.SetNewbies, "новички все").
		AddTriggers("+").
		WithGuards(groupGuard, creatorGuard),
	)

	dp.AddHandler(cf.New(memberHandler.SetOnlyNewbies, "олды кроме").
		AddTriggers("+").
		WithGuards(groupGuard, creatorGuard),
	)

	dp.AddHandler(cf.New(chatHandler.ShowNewbieThreshold, "newbie", "новички", "нью").
		WithGuards(groupGuard, adminGuard),
	)
	dp.AddHandler(cf.New(chatHandler.SetNewbieThreshold, "newbie", "новички", "нью").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)

	dp.AddHandler(cf.New(restHandler.Show, "рест", "rest", "рэст").
		FallbackToSender().
		WithGuards(groupGuard, adminGuard).
		AddTriggers("+", ""),
	)
	dp.AddHandler(cf.New(restHandler.Set, "рест", "rest", "рэст").
		FallbackToSender().
		AddTriggers("+", "").
		WithGuards(groupGuard).
		SetArgsCount(1),
	)
	dp.AddHandler(cf.New(restHandler.End, "-рест", "-rest", "-рэст").
		FallbackToSender().
		WithGuards(groupGuard).
		AddTriggers(""),
	)

	dp.AddHandler(cf.New(adminHandler.ListAdmins, "admins", "админы", "админчики", "администраторы", "адмы", "модеры", "mods").
		WithGuards(groupGuard, rateLimiterGuard),
	)
	dp.AddHandler(cf.New(adminHandler.IsAdmin, "админ", "admin", "is_admin", "адм", "модер", "mod", "is_mod").
		WithGuards(groupGuard).
		FallbackToSender(),
	)
	dp.AddHandler(cf.New(adminHandler.AddAdmin, "+админ", "+admin", "+адм", "+модер", "+mod").
		AddTriggers("").
		WithGuards(groupGuard, adminGuard),
	)
	dp.AddHandler(cf.New(adminHandler.RemoveAdmin, "-администратор", "-админ", "-admin", "-адм", "-модер", "-mod").
		AddTriggers("").
		WithGuards(groupGuard, creatorGuard),
	)

	dp.AddHandler(cf.New(memberHandler.UpdateMembersList, "обновить чат", "update chat", "update").
		WithGuards(groupGuard, rateLimiterGuard),
	)
	dp.AddHandler(cf.New(memberHandler.ListRoles, "роли", "roles", "titles").
		WithGuards(groupGuard, rateLimiterGuard),
	)
	dp.AddHandler(cf.New(memberHandler.DeleteRole, "-роль", "-role", "-title").
		AddTriggers("").
		WithGuards(groupGuard),
	)
	dp.AddHandler(cf.New(memberHandler.ShowRole, "роль", "role", "title",
		"какая роль", "роль у", "роль кого").
		AddTriggers("").
		WithGuards(groupGuard).
		FallbackToSender(),
	)
	dp.AddHandler(cf.New(memberHandler.SetRole, "роль", "role", "title").
		AddTriggers("+").
		WithGuards(groupGuard).
		FallbackToSender().
		SetArgsCount(1),
	)

	dp.AddHandler(cf.New(callHandler.Call, "call", "калл", "колл", "all").
		AddTriggers("+", "").
		WithGuards(groupGuard, adminGuard, rateLimiterGuard).
		SetArgsCount(1),
	)

	dp.AddHandler(cf.New(chatHandler.ShowPrompt, "промпт").AddTriggers(""))
	dp.AddHandler(cf.New(chatHandler.SetPrompt, "промпт").
		AddTriggers("").
		SetArgsCount(1).
		WithGuards(groupGuard, adminGuard),
	)

	dp.AddHandler(cf.New(messageHandler.Bot, "крис").
		AddTriggers("").
		WithGuards(groupGuard, guard.NewRateLimiter(rdb, 5, 10*time.Second)).
		SetArgsCount(1),
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
