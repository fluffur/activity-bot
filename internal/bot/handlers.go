package bot

import (
	"activity-bot/internal/adapter"
	"activity-bot/internal/admin"
	adminH "activity-bot/internal/admin/handler"
	"activity-bot/internal/call"
	callH "activity-bot/internal/call/handler"
	"activity-bot/internal/chat"
	chatH "activity-bot/internal/chat/handler"
	"activity-bot/internal/cmd"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/filter"
	"activity-bot/internal/guard"
	helpH "activity-bot/internal/help/handler"
	"activity-bot/internal/helpers"
	"activity-bot/internal/member"
	memberH "activity-bot/internal/member/handler"
	msg "activity-bot/internal/message"
	messageH "activity-bot/internal/message/handler"
	"activity-bot/internal/rest"
	restH "activity-bot/internal/rest/handler"
	"activity-bot/internal/stats"
	statsH "activity-bot/internal/stats/handler"
	"activity-bot/internal/user"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/chatmember"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

func (a *App) RegisterHandlers() {
	queries := db.New(a.Pool)

	statsRepository := postgres.NewStatsRepository(queries)
	restRepository := postgres.NewRestRepository(queries, a.Pool)
	memberRepository := postgres.NewMemberRepository(queries)
	userRepository := postgres.NewUserRepository(queries)
	chatRepository := postgres.NewChatRepository(queries)
	adminRepository := postgres.NewAdminRepository(queries)
	messageRepository := postgres.NewMessageRepository(queries)

	statsService := stats.NewService(statsRepository)
	restService := rest.NewService(restRepository)
	chatService := chat.NewService(chatRepository, a.Config.DefaultWeeklyNorm)
	userService := user.NewService(userRepository)

	adminsProvider := adapter.NewTelegramChatAdminsProvider(a.Bot)
	statusProvider := adapter.NewTelegramMemberStatusProvider(a.Bot)
	memberService := member.NewService(memberRepository, chatRepository, userRepository, adminsProvider, a.Config.DefaultWeeklyNorm)
	adminService := admin.NewService(adminRepository, statusProvider, a.Config.BotOwnerID)
	messageService := msg.NewService(messageRepository)
	callService := call.NewService(chatRepository, memberService)

	dateParser := helpers.NewDateParser()

	helpHandler := helpH.New(a.Config.BotOwnerID)
	statsHandler := statsH.New(statsService, restService, memberService)
	chatHandler := chatH.New(chatService, adminService, dateParser)
	restHandler := restH.New(restService, userService, adminService, dateParser)
	adminHandler := adminH.New(adminService, userService, memberService)
	messageHandler := messageH.New(messageService, memberService, chatService, a.Deepseek)
	memberHandler := memberH.New(memberService, chatService, userService, callService)
	callHandler := callH.New(callService, chatService)

	adminGuard := guard.NewAdminGuard(adminService)
	creatorGuard := guard.NewCreatorGuard(adminService)
	groupGuard := guard.OnlyGroups()
	rateLimiterGuard := guard.NewRateLimiter(a.Rdb, 1, 10*time.Second)

	cf := cmd.NewFactory(userService, "/", "!", ".")

	a.Dispatcher.AddHandler(cf.New(helpHandler.Start, "start"))
	a.Dispatcher.AddHandler(cf.New(helpHandler.Help, "help"))
	a.Dispatcher.AddHandler(cf.New(callHandler.ShowWelcomeCallMessage, "call_message", "call сообщение", "колл сообщение", "калл сообщение").
		AddTriggers("+", "").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.SetWelcomeCallMessage, "call_message", "call сообщение", "колл сообщение", "калл сообщение").
		AddTriggers("+", "").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ShowStats, "stats", "отчёт", "отчет").
		AddTriggers("").
		SetArgsCount(1).
		WithGuards(groupGuard, rateLimiterGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.WhoAmI, "whoami", "ктоя", "я кто").
		AddTriggers("").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.WhoAmI, "я", "me").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.WhoAreYou, "whoareu", "ктоты", "тыкто").
		AddTriggers("").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.WhoAreYou, "кто", "ты", "you").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.Inactive, "inactive", "неактив", "инактив").
		AddTriggers("").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowNorm, "norm", "норма какая", "а норма какая", "норма", "норма?", "quota", "какая норма", "а какая норма").
		AddTriggers("").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetNorm, "norm", "норма", "quota").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.SetNewbies, "новички все").
		AddTriggers("+").
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.SetOnlyNewbies, "олды кроме").
		AddTriggers("+").
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowNewbieThreshold, "newbie", "новички", "нью").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetNewbieThreshold, "newbie", "новички", "нью").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(restHandler.Show, "рест", "rest", "рэст").
		FallbackToSender().
		WithGuards(groupGuard).
		AddTriggers("+", ""),
	)
	a.Dispatcher.AddHandler(cf.New(restHandler.Set, "рест", "rest", "рэст").
		FallbackToSender().
		AddTriggers("+", "").
		WithGuards(groupGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(restHandler.End, "-рест", "-rest", "-рэст").
		FallbackToSender().
		WithGuards(groupGuard).
		AddTriggers(""),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ListAdmins, "admins", "админы", "админчики", "администраторы", "адмы", "модеры", "mods").
		WithGuards(groupGuard, rateLimiterGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.IsAdmin, "админ", "admin", "is_admin", "адм", "модер", "mod", "is_mod").
		WithGuards(groupGuard).
		FallbackToSender(),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.AddAdmin, "+админ", "+admin", "+адм", "+модер", "+mod").
		AddTriggers("").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.RemoveAdmin, "-администратор", "-админ", "-admin", "-адм", "-модер", "-mod").
		AddTriggers("").
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.UpdateMembersList, "обновить чат", "update chat", "update").
		WithGuards(groupGuard, rateLimiterGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.ListRoles, "роли", "roles", "titles").
		WithGuards(groupGuard, rateLimiterGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.DeleteRole, "-роль", "-role", "-title").
		AddTriggers("").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.ShowRole, "роль", "role", "title",
		"какая роль", "роль у", "роль кого").
		AddTriggers("").
		WithGuards(groupGuard).
		FallbackToSender(),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.SetRole, "роль", "role", "title").
		AddTriggers("+").
		WithGuards(groupGuard).
		FallbackToSender().
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.Call, "call", "калл", "колл", "all").
		AddTriggers("+", "").
		WithGuards(groupGuard, adminGuard, rateLimiterGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowPrompt, "промпт").AddTriggers(""))
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetPrompt, "промпт").
		AddTriggers("+", "").
		SetArgsCount(1).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.EnableCallOnJoin, "call_enable", "включить call", "включить колл", "включить калл").
		AddTriggers("").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.DisableCallOnJoin, "call_disable", "отключить call", "отключить колл", "отключить калл").
		AddTriggers("").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(messageHandler.Bot, "крис").
		AddTriggers("").
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 5, 10*time.Second)).
		SetArgsCount(1),
	)

	a.Dispatcher.AddHandler(handlers.NewMessage(message.LeftChatMember, memberHandler.OnLeftMember))
	a.Dispatcher.AddHandler(handlers.NewMessage(message.NewChatMembers, memberHandler.OnJoinMember))
	a.Dispatcher.AddHandler(handlers.NewMyChatMember(chatmember.NewStatus("administrator"), memberHandler.OnBotPromote))
	a.Dispatcher.AddHandler(handlers.NewMessage(filter.OnlyGroups, messageHandler.Message))
}
