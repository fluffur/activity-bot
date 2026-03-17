package bot

import (
	adminH "activity-bot/internal/admin/handler"
	"activity-bot/internal/call"
	callH "activity-bot/internal/call/handler"
	channelH "activity-bot/internal/channel/handler"
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
	"activity-bot/internal/session"
	"activity-bot/internal/stats"
	statsH "activity-bot/internal/stats/handler"
	userH "activity-bot/internal/user/handler"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/conversation"
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
	sessionRepository := postgres.NewSessionRepository(queries)

	statsService := stats.NewService(statsRepository)
	restService := rest.NewService(restRepository)
	messageService := msg.NewService(messageRepository)

	callService := call.NewService(chatRepository, a.MemberService, statsService)
	sessionService := session.NewService(sessionRepository)

	dateParser := helpers.NewDateParser()

	helpHandler := helpH.New(a.Config.BotOwnerUsername, a.Config.CommandsLink)
	statsHandler := statsH.New(statsService, restService, a.MemberService, a.UserService, a.ChatService, sessionService)
	chatHandler := chatH.New(a.ChatService, a.AdminService, sessionService, dateParser)
	restHandler := restH.New(restService, a.UserService, a.MemberService, a.AdminService, dateParser, sessionService, a.AsyncClient)

	adminHandler := adminH.New(a.AdminService, a.UserService, a.MemberService, a.ChatService, dateParser, a.AsyncClient)

	messageHandler := messageH.New(messageService, a.MemberService, a.ChatService, a.Deepseek)
	memberHandler := memberH.New(a.MemberService, a.ChatService, a.UserService, callService, a.AdminService)
	callHandler := callH.New(callService, a.MemberService, a.ChatService, a.AdminService, sessionService)
	userHandler := userH.New(a.UserService)
	channelHandler := channelH.New(a.AdminService, a.ChatService, a.AsyncClient, a.Config.ChannelID)

	adminGuard := guard.NewAdminGuard(a.AdminService, sessionService)
	creatorGuard := guard.NewCreatorGuard(a.AdminService, sessionService)
	ownerGuard := guard.NewDevCreatorGuard(a.AdminService, sessionService)
	developerGuard := guard.NewDeveloperGuard(a.AdminService, sessionService)
	groupGuard := guard.OnlyGroups(sessionService)
	rateLimiterGuard := guard.NewRateLimiter(a.Rdb, 2, 10*time.Second, sessionService)

	cf := cmd.NewFactory(a.UserService, a.ChatService, a.MemberService, sessionService, a.Config.UniquePrefix, "/", "!", ".")

	a.Dispatcher.AddHandler(cf.New(helpHandler.Start, "start"))
	a.Dispatcher.AddHandler(cf.New(helpHandler.Help, "help"))
	a.Dispatcher.AddHandler(cf.New(callHandler.ShowWelcomeCallMessage, "call_message", "call сообщение", "колл сообщение", "калл сообщение").
		AddTriggers("+").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.SetWelcomeCallMessage, "call_message", "call сообщение", "колл сообщение", "калл сообщение").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)

	a.Dispatcher.AddHandler(cf.New(callHandler.DeleteWelcomeCallMessage, "-call_message", "-call сообщение", "-колл сообщение", "-калл сообщение").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowNewbieThreshold, "newbie", "новички срок", "новички после").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetNewbieThreshold, "newbie", "новички срок", "новички после").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ShowStats, "stats", "отчёт", "отчет", "стата").
		SetArgsCount(1).
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 2, 4*time.Second, sessionService)),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ShowChatActivityGraph, "stats_graph", "график", "граф").
		SetArgsCount(1).
		WithGuards(groupGuard, rateLimiterGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.WhoAreYou, "whoareu", "ктоты", "кто ты", "профиль", "ты кто", "тыкто").
		SetArgsCount(1).
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.WhoAmI, "whoami", "кто я", "профиль", "ктоя", "я кто").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.WhoAmI, "я", "me").ForcePrefix().
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("whoareyou:"), cf.WrapCallback(statsHandler.CallbackWhoAreYou)))
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("profile_graph:"), cf.WrapCallback(statsHandler.CallbackProfileGraph)))
	a.Dispatcher.AddHandler(cf.New(statsHandler.WhoAreYou, "ты", "you").SetArgsCount(1).ForcePrefix().
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ListInactive, "inactive", "неактив", "инактив").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ShowRestList, "rests", "ресты").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(restHandler.AllUserRests, "all_rests", "все ресты", "история рестов").
		WithGuards(groupGuard).FallbackToSender(),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ShowFailedNorm, "nonorm", "без нормы").
		SetArgsCount(1).
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 2, 4*time.Second, sessionService)),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ShowNewbies, "newbies", "новички").
		SetArgsCount(1).
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 2, 4*time.Second, sessionService)),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowNorm, "norm", "норма какая", "а норма какая", "норма", "норма?", "quota", "какая норма", "а какая норма").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetNorm, "norm", "норма", "quota").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(2),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.RemoveNorm, "-norm", "-норма", "-quota").
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
	a.Dispatcher.AddHandler(cf.New(restHandler.Show, "рест", "rest", "мой рест").
		FallbackToSender().
		WithGuards(groupGuard).
		AddTriggers("+"),
	)
	a.Dispatcher.AddHandler(cf.New(restHandler.Set, "рест", "rest", "установить рест").
		FallbackToSender().
		AddTriggers("+").
		WithGuards(groupGuard).
		SetArgsCount(2),
	)
	a.Dispatcher.AddHandler(cf.New(restHandler.End, "-рест", "-rest", "снять рест").
		FallbackToSender().
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ListAdmins, "admins", "админы", "админчики", "администраторы", "адмы", "модеры", "mods").
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 1, 10*time.Second, sessionService)),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.IsAdmin, "админ", "admin", "is_admin", "адм", "модер", "mod", "is_mod").
		WithGuards(groupGuard).
		FallbackToSender(),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.AddAdmin, "+админ", "+admin", "+адм", "+модер", "+mod").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.RemoveAdmin, "-администратор", "-админ", "-admin", "-адм", "-модер", "-mod").
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Unban, "unban", "-бан", "разбан", "разбанить").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Unmute, "unmute", "размут", "размутить", "-мут").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Unwarn, "unwarn", "анварн", "снятьпред", "-варн", "-пред").
		WithGuards(groupGuard, adminGuard),
	)

	a.Dispatcher.AddHandler(cf.New(adminHandler.Kick, "kick", "кик", "выгнать").
		AddTriggers("+").
		SetArgsCount(1).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Ban, "ban", "бан").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(2),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Mute, "mute", "мут", "замутить").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(2),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ShowWarns, "warns", "варны", "преды").
		FallbackToSender().
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Warnlist, "warnlist", "варнлист", "предывсе").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Warn, "warn", "варн", "пред", "предупреждение").
		AddTriggers("+").SetArgsCount(2).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ClearWarns, "clear_warns", "очистить преды", "очистить варны").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ShowMaxWarns, "макс преды", "макс варны", "max_warns").
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
		WithGuards(ownerGuard).SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ListDevelopers, "девс", "devs").
		WithGuards(developerGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.UpdateChats, "update_chats").
		WithGuards(developerGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.UpdateMembersList, "обновить чат", "update chat", "update").
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 1, 10*time.Second, sessionService)),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.ListRoles, "роли", "roles", "titles").
		WithGuards(groupGuard, rateLimiterGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.ShowRole, "роль", "role", "title",
		"какая роль", "роль у", "роль кого").
		WithGuards(groupGuard).
		FallbackToSender(),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.SetRole, "роль", "role", "title").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.RestoreRoles, "восстановить роли", "restore_roles").
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.ShowCallTypes, "call_type", "калл тип", "калл стиль").
		AddTriggers("+", "!").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.SetMentionsPerMessage, "call_limit", "калл лимит", "калл лим").
		AddTriggers("+", "!").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Equal("call_style"), cf.WrapCallback(callHandler.ShowCallTypes)))

	a.Dispatcher.AddHandler(cf.New(callHandler.CallInactive, "call_inactive", "калл инактив", "калл неактив", "созвать неактивных").
		WithGuards(groupGuard, adminGuard, rateLimiterGuard),
	)

	a.Dispatcher.AddHandler(cf.New(callHandler.CallNoNorm, "call_no_norm", "калл без нормы", "созвать без нормы").
		WithGuards(groupGuard, adminGuard, rateLimiterGuard),
	)

	a.Dispatcher.AddHandler(cf.New(callHandler.CallNoNormWarn, "call_no_norm_warn", "калл без нормы варн", "созвать без нормы варн").
		WithGuards(groupGuard, adminGuard, rateLimiterGuard),
	)

	a.Dispatcher.AddHandler(cf.New(callHandler.CallNoNormWarn, "call_no_norm_ban", "калл без нормы бан", "созвать без нормы бан").
		WithGuards(groupGuard, adminGuard, rateLimiterGuard),
	)
	callConversation := handlers.NewConversation(
		[]ext.Handler{
			handlers.NewCallback(callbackquery.Equal("call_inactive"), callHandler.StartCallInactiveConversation),
			handlers.NewCallback(callbackquery.Equal("call_no_norm_warn"), callHandler.StartCallNoNormWarnConversation),
			handlers.NewCallback(callbackquery.Equal("call_no_norm_ban"), callHandler.StartCallNoNormBanConversation),
			handlers.NewCallback(callbackquery.Equal("call_no_norm"), callHandler.StartCallNoNormConversation),
		},
		map[string][]ext.Handler{
			callH.CallStateInactive: {
				handlers.NewMessage(message.Text, callHandler.HandleCallInactiveMessage),
			},
			callH.CallStateNoNorm: {
				handlers.NewMessage(message.Text, callHandler.HandleCallNoNormMessage),
			},
			callH.CallStateNoNormWarn: {
				handlers.NewMessage(message.Text, callHandler.HandleCallNoNormWarnMessage),
			},
			callH.CallStateNoNormBan: {
				handlers.NewMessage(message.Text, callHandler.HandleCallNoNormBanMessage),
			},
		},
		&handlers.ConversationOpts{
			Exits: []ext.Handler{
				handlers.NewCallback(callbackquery.Prefix("call_cancel"), callHandler.CancelCallConversation),
				handlers.NewCallback(callbackquery.Prefix("call_nomsg:"), callHandler.NoMessageCallConversation),
			},
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
			AllowReEntry: true,
		},
	)
	a.Dispatcher.AddHandler(callConversation)

	a.Dispatcher.AddHandler(cf.New(callHandler.Call, "call", "калл", "колл", "all", "каллалл").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard, rateLimiterGuard).
		SetArgsCount(1),
	)

	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("call_type:"), cf.WrapCallback(callHandler.CallbackCallType)))
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowPrompt, "промпт").WithGuards(groupGuard))
	a.Dispatcher.AddHandler(handlers.NewMessage(cmd.NewChatTitle, cf.WrapEvent(chatHandler.OnNewChatTitle)))
	a.Dispatcher.AddHandler(cf.New(chatHandler.Manage, "manage", "управление"))
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("manage:"), cf.WrapCallback(chatHandler.CallbackManage)))
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("manage_page:"), cf.WrapCallback(chatHandler.CallbackManagePage)))
	a.Dispatcher.AddHandler(cf.New(chatHandler.EnableTags, "+tags", "+теги", "+тэги").WithGuards(groupGuard, adminGuard))
	a.Dispatcher.AddHandler(cf.New(chatHandler.DisableTags, "-tags", "-теги", "-тэги").WithGuards(groupGuard, adminGuard))
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowTags, "tags", "теги", "тэги"))
	a.Dispatcher.AddHandler(cf.New(chatHandler.UserChats, "chats", "чаты", "нормы", "чаты без нормы"))
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetPrompt, "промпт").
		AddTriggers("+").
		SetArgsCount(1).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowWeekStart, "week_start", "начало недели", "чистка", "время чистки", "конец чистки"))
	a.Dispatcher.AddHandler(cf.New(chatHandler.ManageWeekStart, "week_start", "начало недели", "чистка", "время чистки", "конец чистки").
		AddTriggers("+").
		SetArgsCount(1).WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowPrefix, "custom_prefix", "кастом префикс", "префикс").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetPrefix, "custom_prefix", "кастом префикс", "префикс").
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowPrefixes, "префиксы", "prefixes").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.DisablePrefixes, "с префиксами", "-prefixless").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.EnablePrefixes, "без префиксов", "+prefixless").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.EnableCallOnJoin, "call_enable", "включить call", "включить колл", "включить калл").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.DisableCallOnJoin, "call_disable", "отключить call", "отключить колл", "отключить калл").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(messageHandler.Bot, "крис").
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 5, 10*time.Second, sessionService)).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.DemoteTgAdmin, "разжаловать").WithGuards(groupGuard, adminGuard))
	a.Dispatcher.AddHandler(cf.New(adminHandler.FakeLeave, "фейклив").FallbackToSender().WithGuards(groupGuard))
	a.Dispatcher.AddHandler(cf.New(userHandler.ShowGender, "пол", "gender").FallbackToSender())
	a.Dispatcher.AddHandler(cf.New(userHandler.SetGender, "пол", "gender").
		FallbackToSender().
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(userHandler.ShowEmoji, "эмоджи", "эмодзи").FallbackToSender())
	a.Dispatcher.AddHandler(cf.New(userHandler.SetEmoji, "эмоджи", "эмодзи").FallbackToSender().SetArgsCount(1))

	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("approve:"), cf.WrapCallback(restHandler.ApproveRestRequest)))
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("reject:"), cf.WrapCallback(restHandler.RejectRestRequest)))

	a.Dispatcher.AddHandler(handlers.NewMessage(message.LeftChatMember, cf.WrapEvent(memberHandler.OnLeftMember)))
	a.Dispatcher.AddHandler(handlers.NewMyChatMember(chatmember.NewStatus("administrator"), cf.WrapEvent(memberHandler.OnBotPromote)))
	a.Dispatcher.AddHandler(handlers.NewMessage(message.NewChatMembers, cf.WrapEvent(memberHandler.OnJoinMember)))
	a.Dispatcher.AddHandler(handlers.NewMessage(message.Channel, cf.WrapEvent(channelHandler.Post)).SetAllowChannel(true))
	a.Dispatcher.AddHandler(cf.New(channelHandler.Subscribe, "subscribe").WithGuards(groupGuard, adminGuard))
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("unsubscribe"), cf.WrapEvent(channelHandler.Unsubscribe)))
	a.Dispatcher.AddHandler(handlers.NewMessage(filter.OnlyGroups, cf.WrapEvent(messageHandler.Message)))

}
