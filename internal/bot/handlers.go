package bot

import (
	adminH "activity-bot/internal/admin/handler"
	"activity-bot/internal/call"
	callH "activity-bot/internal/call/handler"
	chatH "activity-bot/internal/chat/handler"
	"activity-bot/internal/cmd"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/filter"
	"activity-bot/internal/guard"
	helpH "activity-bot/internal/help/handler"
	"activity-bot/internal/helpers"
	memberH "activity-bot/internal/member/handler"
	msg "activity-bot/internal/message"
	messageH "activity-bot/internal/message/handler"
	"activity-bot/internal/rest"
	restH "activity-bot/internal/rest/handler"
	"activity-bot/internal/stats"
	statsH "activity-bot/internal/stats/handler"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/chatmember"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
)

func (a *App) RegisterHandlers() {
	queries := db.New(a.Pool)

	statsRepository := postgres.NewStatsRepository(queries)
	restRepository := postgres.NewRestRepository(queries, a.Pool)
	chatRepository := postgres.NewChatRepository(queries)
	messageRepository := postgres.NewMessageRepository(queries)

	statsService := stats.NewService(statsRepository)
	restService := rest.NewService(restRepository)
	messageService := msg.NewService(messageRepository)

	callService := call.NewService(chatRepository, a.MemberService)

	dateParser := helpers.NewDateParser()

	helpHandler := helpH.New(a.Config.BotOwnerID)
	statsHandler := statsH.New(statsService, restService, a.MemberService, a.UserService, a.ChatService)
	chatHandler := chatH.New(a.ChatService, a.AdminService, dateParser)
	restHandler := restH.New(restService, a.UserService, a.AdminService, dateParser)

	adminHandler := adminH.New(a.AdminService, a.UserService, a.MemberService, dateParser, a.AsyncClient)

	messageHandler := messageH.New(messageService, a.MemberService, a.ChatService, a.Deepseek)
	memberHandler := memberH.New(a.MemberService, a.ChatService, a.UserService, callService)
	callHandler := callH.New(callService, a.ChatService)

	adminGuard := guard.NewAdminGuard(a.AdminService)
	creatorGuard := guard.NewCreatorGuard(a.AdminService)
	ownerGuard := guard.NewDevCreatorGuard(a.AdminService)
	developerGuard := guard.NewDeveloperGuard(a.AdminService)
	groupGuard := guard.OnlyGroups()
	rateLimiterGuard := guard.NewRateLimiter(a.Rdb, 1, 10*time.Second)
	moderationGuard := guard.NewModerationGuard(a.ChatService)

	cf := cmd.NewFactory(a.UserService, a.ChatService, a.Config.UniquePrefix, "/", "!", ".")

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
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 1, 4*time.Second)),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ShowChatActivityGraph, "stats_graph", "график", "граф").
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
		AddTriggers("").SetArgsCount(1).
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("whoareyou:"), statsHandler.CallbackWhoAreYou))
	a.Dispatcher.AddHandler(cf.New(statsHandler.WhoAreYou, "кто", "ты", "you").SetArgsCount(1).
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
		SetArgsCount(2),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.SetNewbies, "новички все").
		AddTriggers("+").
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.SetOnlyNewbies, "олды кроме").
		AddTriggers("+").
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowNewbieThreshold, "newbie", "новички", "новички после").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetNewbieThreshold, "newbie", "новички", "новички после").
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
	a.Dispatcher.AddHandler(cf.New(adminHandler.Unban, "unban", "-бан", "разбан", "разбанить").
		AddTriggers("").
		WithGuards(groupGuard, adminGuard, moderationGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Unmute, "unmute", "размут", "размутить", "-мут").
		AddTriggers("").
		WithGuards(groupGuard, adminGuard, moderationGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Unwarn, "unwarn", "анварн", "снятьпред", "-варн", "-пред").
		AddTriggers("").
		WithGuards(groupGuard, adminGuard, moderationGuard),
	)

	a.Dispatcher.AddHandler(cf.New(adminHandler.Kick, "kick", "кик", "выгнать").
		AddTriggers("", "+").
		SetArgsCount(1).
		WithGuards(groupGuard, adminGuard, moderationGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Ban, "ban", "бан").
		AddTriggers("", "+").
		WithGuards(groupGuard, adminGuard, moderationGuard).
		SetArgsCount(2),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Mute, "mute", "мут", "замутить").
		AddTriggers("", "+").
		WithGuards(groupGuard, adminGuard, moderationGuard).
		SetArgsCount(2),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ShowWarns, "warns", "варны", "преды").
		AddTriggers("").FallbackToSender().
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Warn, "warn", "варн", "пред", "предупреждение").
		AddTriggers("", "+").SetArgsCount(2).
		WithGuards(groupGuard, adminGuard, moderationGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ClearWarns, "clear_warns", "очистить преды").
		WithGuards(groupGuard, adminGuard, moderationGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ShowMaxWarns, "макс преды", "макс варны", "max_warns").
		AddTriggers("").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.SetMaxWarns, "max_warns", "макс преды", "макс варны").
		AddTriggers("+").
		SetArgsCount(1).
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ToggleRights, "права", "rights").
		WithGuards(developerGuard).SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.AddDeveloper, "дев", "adddev").
		WithGuards(ownerGuard).SetArgsCount(1).
		AddTriggers("+"),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.RemoveDeveloper, "-дев", "remdev").
		AddTriggers("").
		WithGuards(ownerGuard).SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ListDevelopers, "девс", "devs").
		WithGuards(developerGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.UpdateMembersList, "обновить чат", "update chat", "update").
		WithGuards(groupGuard, rateLimiterGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.ListRoles, "роли", "roles", "titles").
		WithGuards(groupGuard, rateLimiterGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.DeleteRole, "-роль", "-role", "-title").
		AddTriggers("").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.ShowRole, "роль", "role", "title",
		"какая роль", "роль у", "роль кого").
		AddTriggers("").
		WithGuards(groupGuard).
		FallbackToSender(),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.SetRole, "роль", "role", "title").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		FallbackToSender().
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.RestoreRoles, "восстановить роли", "restore_roles").
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.Call, "call", "калл", "колл", "all", "каллалл").
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
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetWeekStartDay, "week_start", "начало недели", "начало").
		AddTriggers("+", "").
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowPrefix, "prefix", "префикс").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetPrefix, "prefix", "префикс").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetModerationEnabled, "модерация", "moderation").
		AddTriggers("+", "-").
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
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("approve:"), restHandler.ApproveRestRequest))
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("reject:"), restHandler.RejectRestRequest))
	a.Dispatcher.AddHandler(handlers.NewMessage(message.LeftChatMember, memberHandler.OnLeftMember))
	a.Dispatcher.AddHandler(handlers.NewMessage(message.NewChatMembers, memberHandler.OnJoinMember))
	a.Dispatcher.AddHandler(handlers.NewMyChatMember(chatmember.NewStatus("administrator"), memberHandler.OnBotPromote))
	a.Dispatcher.AddHandler(handlers.NewMessage(filter.OnlyGroups, messageHandler.Message))
}
