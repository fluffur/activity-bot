package bot

import (
	adminH "activity-bot/internal/admin/handler"
	"activity-bot/internal/call"
	callH "activity-bot/internal/call/handler"
	channelH "activity-bot/internal/channel/handler"
	chatH "activity-bot/internal/chat/handler"
	"activity-bot/internal/cmd"
	"activity-bot/internal/command"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/filter"
	"activity-bot/internal/guard"
	helpH "activity-bot/internal/help/handler"
	"activity-bot/internal/helpers"
	memberH "activity-bot/internal/member/handler"
	msg "activity-bot/internal/message"
	messageH "activity-bot/internal/message/handler"
	"activity-bot/internal/model"
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

	cf := cmd.NewFactory(a.UserService, a.ChatService, a.MemberService, sessionService, a.Config.UniquePrefix, "/", "!", ".")

	helpHandler := helpH.New(a.Config.BotOwnerUsername, a.Config.CommandsLink)
	statsHandler := statsH.New(statsService, restService, a.MemberService, a.UserService, a.ChatService, sessionService)
	chatHandler := chatH.New(a.ChatService, a.AdminService, a.MemberService, sessionService, dateParser)
	restHandler := restH.New(restService, a.UserService, a.MemberService, a.ChatService, a.AdminService, dateParser, sessionService, a.AsyncClient)

	adminHandler := adminH.New(a.AdminService, a.MemberService, a.ChatService, dateParser, a.AsyncClient, cf)

	messageHandler := messageH.New(messageService, a.MemberService, a.ChatService, a.Deepseek)
	memberHandler := memberH.New(a.MemberService, a.ChatService, a.UserService, callService, a.AdminService)
	callHandler := callH.New(callService, a.MemberService, a.ChatService, a.AdminService, sessionService)
	userHandler := userH.New(a.UserService)
	channelHandler := channelH.New(a.MemberService, a.ChatService, a.AsyncClient, a.Config.ChannelID)

	developerGuard := guard.NewDeveloperGuard(a.AdminService, a.Config.BotOwnerID)
	groupGuard := guard.OnlyGroups(sessionService)
	rateLimiterGuard := guard.NewRateLimiter(a.Rdb, 2, 10*time.Second, sessionService)

	f := command.NewCommandFactory(a.UserService, a.MemberService, a.ChatService, sessionService, "фм", "!", "/", ".")
	a.Dp.AddHandler(f.New("start", helpHandler.Start).SetScope(command.ScopeUser))
	a.Dp.AddHandler(f.New("help", helpHandler.Help).SetScope(command.ScopeUser))
	a.Dp.AddHandler(f.New("show_call_message", callHandler.ShowWelcomeCallMessage).
		SetAliases("калл сообщение").
		SetProviders(a.UserService, a.MemberService, a.ChatService, sessionService),
	)
	a.Dp.AddHandler(f.New("set_call_message", callHandler.SetWelcomeCallMessage).
		SetAliases("калл сообщение").
		AddTriggers("+").
		SetDescription("Настройка сообщения сбора").
		SetCategory(cmd.CategoryCall).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.TextRule()),
	)
	a.Dp.AddHandler(f.New("delete_call_message", callHandler.DeleteWelcomeCallMessage).
		SetAliases("калл сообщение").AddTriggers("-").
		SetRequiredStatus(model.StatusSeniorAdmin),
	)
	a.Dp.AddHandler(f.New("show_newbie_treshold", chatHandler.ShowNewbieThreshold).
		SetAliases("новички срок", "новички после").
		SetRequiredStatus(model.StatusSeniorAdmin),
	)
	a.Dp.AddHandler(f.New("set_newbie_treshold", chatHandler.SetNewbieThreshold).
		SetAliases("новички срок", "новички после").
		AddTriggers("+").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.NumberRule()).
		SetDescription("Настройка срока новичка").
		SetCategory(cmd.CategorySettings),
	)
	a.Dp.AddHandler(f.New("stats", statsHandler.ShowStats).
		SetAliases("отчёт", "отчет", "стата").
		SetArgRules(command.OptionalDateRangeRule()).
		//WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 2, 4*time.Second, sessionService)).
		SetDescription("Отчёт в чате").
		SetCategory(cmd.CategoryStats),
	)
	a.Dp.AddHandler(f.New("stats_graph", statsHandler.ShowChatActivityGraph).
		SetAliases("график").
		SetArgRules(command.OptionalDateRangeRule()),
	//WithGuards(groupGuard, rateLimiterGuard),
	)

	a.Dp.AddHandler(f.New("who_am_i", statsHandler.WhoAmI).
		SetAliases("ктоя", "кто я", "профиль").SetArgRules(command.ArgRule{
		Name: "only_sender",
		Type: command.ArgTypeOnlyUserSender,
	}))

	a.Dp.AddHandler(f.New("who_are_you", statsHandler.WhoAreYou).
		SetAliases("ктоты", "кто ты", "профиль").
		SetArgRules(command.AnyUserRule()))

	a.Dp.AddHandler(f.New("inactive", statsHandler.ListInactive).
		SetAliases("инактив", "неактив"))

	a.Dp.AddHandler(f.New("rests", statsHandler.ShowRestList).
		SetAliases("ресты"),
	)

	a.Dp.AddHandler(f.New("all_rests", restHandler.AllUserRests).
		SetAliases("все ресты").SetArgRules(command.AnyUserRule()),
	)

	a.Dp.AddHandler(f.New("failed_norm", statsHandler.ShowFailedNorm).
		SetArgRules(command.OptionalDateRangeRule()),
	)

	a.Dp.AddHandler(f.New("newbies", statsHandler.ShowNewbies).
		SetAliases("новички"),
	)

	a.Dp.AddHandler(f.New("norm", chatHandler.ShowNorm).
		SetAliases("норма какая", "а норма какая", "норма", "норма?", "quota", "какая норма", "а какая норма"),
	)
	a.Dp.AddHandler(f.New("set_norm", chatHandler.SetNorm).SetAliases("норма", "quota").
		AddTriggers("+").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.NumberRule(), command.TextRule().SetVariadic(true).SetRange(0, 1)),
	)
	a.Dp.AddHandler(f.New("remove_norm", chatHandler.RemoveNorm).SetAliases("норма", "quota").
		AddTriggers("-").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.TextRule().SetVariadic(true).SetRange(0, 1)),
	)
	a.Dp.AddHandler(f.New("ship_random", memberHandler.ShipRandom).
		SetAliases("рандом шипперим"))

	a.Dp.AddHandler(f.New("set_rest", restHandler.SetRest).SetRequiredStatus(model.StatusModerator).
		SetAliases("рест", "rest", "установить рест").
		AddTriggers("+").
		SetArgRules(command.AnyUserRule(), command.OneDateRule()),
	)
	a.Dp.AddHandler(f.New("rest", restHandler.ShowRest).
		SetAliases("рест", "rest").
		SetArgRules(command.AnyUserRule()),
	)
	a.Dp.AddHandler(f.New("remove_rest", restHandler.RemoveRestRequest).
		SetAliases("удалить рест").
		SetRequiredStatus(model.StatusAdmin).
		SetArgRules(command.AnyUserRule(), command.NumberRule()),
	)
	a.Dp.AddHandler(f.New("end_rest", restHandler.EndRest).SetAliases("рест", "rest", "снять рест").
		AddTriggers("-").
		SetRequiredStatus(model.StatusModerator).
		SetArgRules(command.AnyUserRule()),
	)
	a.Dp.AddHandler(f.New("admins", adminHandler.ListAdmins).
		SetAliases("админы", "модеры", "mods"),
	)

	a.Dp.AddHandler(f.New("add_admin", adminHandler.SetStatus).SetAliases("админ", "admin", "mod").
		SetArgRules(command.AnyUserRule(), command.NumberRule()).AddTriggers("+").
		SetRequiredStatus(model.StatusCoOwner),
	)
	a.Dp.AddHandler(f.New("is_admin", adminHandler.IsAdmin).SetAliases("админ", "admin", "mod").
		SetArgRules(command.AnyUserRule()),
	)
	a.Dp.AddHandler(f.New("remove_admin", adminHandler.RemoveAdmin).SetAliases("админ", "admin", "mod").
		SetTriggers("-").SetArgRules(command.MentionedUserRule()).
		SetRequiredStatus(model.StatusCoOwner),
	)
	a.Dp.AddHandler(cf.New(adminHandler.Unban, "unban", "-бан", "разбан", "разбанить").
		WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin).
		WithDescription("Разбан участника").
		WithCategory(cmd.CategoryModeration).
		WithAmbiguityResolution("admin_unban"),
	)
	a.Dp.AddHandler(cf.New(adminHandler.Unmute, "unmute", "размут", "размутить", "-мут").
		WithGuards(groupGuard).
		Restricted(model.StatusAdmin).
		WithDescription("Размут участника").
		WithCategory(cmd.CategoryModeration).
		WithAmbiguityResolution("admin_unmute"),
	)
	a.Dp.AddHandler(cf.New(adminHandler.Unwarn, "unwarn", "снять пред", "-варн", "-пред").
		WithGuards(groupGuard).
		Restricted(model.StatusAdmin).
		WithDescription("Снять предупреждение").
		WithCategory(cmd.CategoryModeration).
		WithAmbiguityResolution("admin_unwarn"),
	)

	a.Dp.AddHandler(cf.New(adminHandler.Kick, "kick", "кик", "выгнать").
		AddTriggers("+").
		SetArgsCount(1).
		WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin).
		WithDescription("Кик участника").
		WithCategory(cmd.CategoryModeration).
		WithAmbiguityResolution("admin_kick"),
	)
	a.Dp.AddHandler(cf.New(adminHandler.Ban, "ban", "бан").
		AddTriggers("+").
		WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin).
		SetArgsCount(2).
		WithDescription("Бан участника").
		WithCategory(cmd.CategoryModeration).
		WithAmbiguityResolution("admin_ban"),
	)
	a.Dp.AddHandler(cf.New(adminHandler.Mute, "mute", "мут").
		AddTriggers("+").
		WithGuards(groupGuard).
		Restricted(model.StatusAdmin).
		SetArgsCount(2).
		WithDescription("Мут участника").
		WithCategory(cmd.CategoryModeration).
		WithAmbiguityResolution("admin_mute"),
	)
	a.Dp.AddHandler(cf.New(adminHandler.ShowWarns, "warns", "варны", "преды").
		FallbackToSender().
		WithGuards(groupGuard).
		WithAmbiguityResolution("admin_warns"),
	)
	a.Dp.AddHandler(cf.New(adminHandler.Warnlist, "warnlist", "варнлист", "предывсе").
		WithGuards(groupGuard),
	)
	a.Dp.AddHandler(cf.New(adminHandler.Warn, "warn", "варн", "пред").
		AddTriggers("+").SetArgsCount(2).
		WithGuards(groupGuard).
		Restricted(model.StatusAdmin).
		WithDescription("Предупреждение").
		WithCategory(cmd.CategoryModeration).
		WithAmbiguityResolution("admin_warn"),
	)
	a.Dp.AddHandler(cf.New(adminHandler.ClearWarns, "clear_warns", "очистить преды", "очистить варны").
		WithGuards(groupGuard).
		Restricted(model.StatusAdmin).
		WithDescription("Очистить предупреждения").
		WithCategory(cmd.CategoryModeration).
		WithAmbiguityResolution("admin_clear"),
	)
	a.Dp.AddHandler(cf.New(adminHandler.ShowMaxWarns, "макс преды", "макс варны", "max_warns").
		WithGuards(groupGuard),
	)
	a.Dp.AddHandler(cf.New(adminHandler.SetMaxWarns, "max_warns", "макс преды", "макс варны").
		AddTriggers("+").
		SetArgsCount(1).
		WithGuards(groupGuard).
		Restricted(model.StatusCoOwner),
	)
	a.Dp.AddHandler(cf.New(adminHandler.ToggleRights, "права", "rights").
		WithGuards(groupGuard, developerGuard).SetArgsCount(1).FallbackToSender(),
	)
	a.Dp.AddHandler(cf.New(adminHandler.UpdateChats, "update_chats").
		WithGuards(groupGuard, developerGuard),
	)
	a.Dp.AddHandler(cf.New(memberHandler.UpdateMembersList, "обновить чат", "update chat", "update").
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 1, 30*time.Second, sessionService)).
		Restricted(model.StatusModerator).
		WithDescription("Обновление списка участников").
		WithCategory(cmd.CategorySettings),
	)
	a.Dp.AddHandler(cf.New(memberHandler.ListRoles, "роли", "roles", "titles").
		WithGuards(groupGuard, rateLimiterGuard).Restricted(model.StatusMember).WithDescription("Список ролей (тегов) участников").WithCategory(cmd.CategoryStats),
	)
	a.Dp.AddHandler(cf.New(memberHandler.ShowRole, "роль", "role", "title",
		"какая роль", "роль у", "роль кого").
		WithGuards(groupGuard).
		FallbackToSender().
		WithAmbiguityResolution("member_role_show"),
	)
	a.Dp.AddHandler(cf.New(memberHandler.SetRole, "роль", "role", "title").
		AddTriggers("+").
		WithGuards(groupGuard).
		Restricted(model.StatusModerator).
		SetArgsCount(1).
		WithDescription("Присвоение ролей").
		WithCategory(cmd.CategoryStats).
		WithAmbiguityResolution("member_role_set"),
	)
	a.Dp.AddHandler(cf.New(memberHandler.RestoreRoles, "перенести админки", "move admins").
		WithGuards(groupGuard).
		Restricted(model.StatusModerator).
		WithDescription("Перенос тг админок").
		WithCategory(cmd.CategoryModeration),
	)

	a.Dp.AddHandler(cf.New(callHandler.ShowCallTypes, "call_type", "калл тип", "калл стиль").
		AddTriggers("+", "!").
		WithGuards(groupGuard),
	)
	a.Dp.AddHandler(cf.New(callHandler.SetMentionsPerMessage, "call_limit", "калл лимит", "калл лим").
		AddTriggers("+", "!").
		WithGuards(groupGuard).
		Restricted(model.StatusCoOwner).
		SetArgsCount(1),
	)
	a.Dp.AddHandler(handlers.NewCallback(callbackquery.Equal("call_style"), cf.WrapCallback(callHandler.ShowCallTypes)))

	a.Dp.AddHandler(cf.New(callHandler.CallInactive, "call_inactive", "калл инактив", "калл неактив", "созвать неактивных").
		WithGuards(groupGuard, rateLimiterGuard).
		Restricted(model.StatusModerator).
		WithDescription("Сбор неактивных").
		WithCategory(cmd.CategoryCall),
	)

	a.Dp.AddHandler(cf.New(callHandler.CallNoNorm, "call_no_norm", "калл без нормы", "созвать без нормы").
		WithGuards(groupGuard, rateLimiterGuard).
		Restricted(model.StatusModerator).
		WithDescription("Сбор тех, кто без нормы").
		WithCategory(cmd.CategoryCall),
	)

	a.Dp.AddHandler(cf.New(callHandler.CallNoNormWarn, "call_no_norm_warn", "калл без нормы варн", "созвать без нормы варн").
		WithGuards(groupGuard, rateLimiterGuard).
		Restricted(model.StatusModerator),
	)

	a.Dp.AddHandler(cf.New(callHandler.CallNoNormWarn, "call_no_norm_ban", "калл без нормы бан", "созвать без нормы бан").
		WithGuards(groupGuard, rateLimiterGuard).
		Restricted(model.StatusModerator),
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
	a.Dp.AddHandler(callConversation)

	a.Dp.AddHandler(cf.New(callHandler.Call, "call", "калл", "колл", "all", "каллалл").
		AddTriggers("+").
		WithGuards(groupGuard, rateLimiterGuard).
		Restricted(model.StatusModerator).
		SetArgsCount(1).
		WithDescription("Общий сбор чата").
		WithCategory(cmd.CategoryCall),
	)

	a.Dp.AddHandler(handlers.NewCallback(callbackquery.Prefix("call_type:"), cf.WrapCallback(callHandler.CallbackCallType)))
	a.Dp.AddHandler(cf.New(chatHandler.ShowPrompt, "промпт").WithGuards(groupGuard))
	a.Dp.AddHandler(handlers.NewMessage(cmd.NewChatTitle, cf.WrapEvent(chatHandler.OnNewChatTitle)))
	a.Dp.AddHandler(cf.New(chatHandler.Manage, "manage", "управление"))
	a.Dp.AddHandler(handlers.NewCallback(callbackquery.Prefix("manage:"), cf.WrapCallback(chatHandler.CallbackManage)))
	a.Dp.AddHandler(handlers.NewCallback(callbackquery.Prefix("manage_page:"), cf.WrapCallback(chatHandler.CallbackManagePage)))
	a.Dp.AddHandler(cf.New(chatHandler.EnableTags, "+tags", "+теги", "+тэги").WithGuards(groupGuard).Restricted(model.StatusSeniorAdmin).WithDescription("Включение тегов").WithCategory(cmd.CategorySettings))
	a.Dp.AddHandler(cf.New(chatHandler.DisableTags, "-tags", "-теги", "-тэги").WithGuards(groupGuard).Restricted(model.StatusSeniorAdmin).WithDescription("Отключение тегов").WithCategory(cmd.CategorySettings))
	a.Dp.AddHandler(cf.New(chatHandler.ShowTags, "tags", "теги", "тэги"))
	a.Dp.AddHandler(cf.New(chatHandler.UserChats, "chats", "чаты", "нормы", "чаты без нормы"))
	a.Dp.AddHandler(cf.New(chatHandler.SetPrompt, "промпт").
		AddTriggers("+").
		SetArgsCount(1).
		WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin).
		WithDescription("Настройка промпта для ИИ").
		WithCategory(cmd.CategorySettings),
	)
	a.Dp.AddHandler(cf.New(chatHandler.ShowWeekStart, "week_start", "начало недели", "чистка", "время чистки", "конец чистки"))
	a.Dp.AddHandler(cf.New(chatHandler.ManageWeekStart, "week_start", "начало недели", "чистка", "время чистки", "конец чистки").
		AddTriggers("+").
		SetArgsCount(1).WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin).
		WithDescription("Настройка начала недели").
		WithCategory(cmd.CategorySettings),
	)
	a.Dp.AddHandler(cf.New(chatHandler.ShowPrefix, "custom_prefix", "кастом префикс", "префикс").
		WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin),
	)
	a.Dp.AddHandler(cf.New(chatHandler.SetPrefix, "custom_prefix", "кастом префикс", "префикс").
		AddTriggers("+").
		WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin).
		SetArgsCount(1).
		WithDescription("Кастомные префиксы").
		WithCategory(cmd.CategorySettings),
	)
	a.Dp.AddHandler(cf.New(chatHandler.ShowPrefixes, "префиксы", "prefixes").
		WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin),
	)
	a.Dp.AddHandler(cf.New(chatHandler.DisablePrefixes, "с префиксами", "-prefixless").
		WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin),
	)
	a.Dp.AddHandler(cf.New(chatHandler.EnablePrefixes, "без префиксов", "+prefixless").
		WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin),
	)
	a.Dp.AddHandler(cf.New(callHandler.EnableCallOnJoin, "call_enable", "включить call", "включить колл", "включить калл").
		WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin).
		WithDescription("Включить призыв при входе").
		WithCategory(cmd.CategoryCall),
	)
	a.Dp.AddHandler(cf.New(callHandler.DisableCallOnJoin, "call_disable", "отключить call", "отключить колл", "отключить калл").
		WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin),
	)
	a.Dp.AddHandler(cf.New(messageHandler.Bot, "крис").
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 5, 10*time.Second, sessionService)).
		SetArgsCount(1),
	)
	a.Dp.AddHandler(cf.New(adminHandler.DemoteTgAdmin, "разжаловать").WithGuards(groupGuard).Restricted(model.StatusSeniorAdmin))
	a.Dp.AddHandler(cf.New(adminHandler.FakeLeave, "фейклив").FallbackToSender().WithGuards(groupGuard))
	a.Dp.AddHandler(cf.New(userHandler.ShowGender, "пол", "gender").FallbackToSender())
	a.Dp.AddHandler(cf.New(userHandler.SetGender, "пол", "gender", "установить пол").
		FallbackToSender().
		SetArgsCount(1),
	)
	a.Dp.AddHandler(cf.New(userHandler.ShowEmoji, "эмоджи", "эмодзи").FallbackToSender())
	a.Dp.AddHandler(cf.New(userHandler.RemoveEmoji, "-эмоджи", "-эмодзи").FallbackToSender())
	a.Dp.AddHandler(cf.New(userHandler.SetEmoji, "эмоджи", "эмодзи").FallbackToSender().SetArgsCount(1))

	a.Dp.AddHandler(cf.New(adminHandler.ManageRights, "дк").
		WithGuards(groupGuard).
		Restricted(model.StatusCoOwner).
		WithDescription("Управление доступом команд").
		WithCategory(cmd.CategorySettings),
	)
	a.Dp.AddHandler(handlers.NewCallback(callbackquery.Prefix("rights_"), cf.WrapCallback(adminHandler.CallbackManageRights, guard.Restricted(a.MemberService, a.ChatService, sessionService, "дк", model.StatusSeniorAdmin))))

	a.Dp.AddHandler(handlers.NewCallback(callbackquery.Prefix("approve:"), cf.WrapCallback(restHandler.ApproveRestRequest, guard.Restricted(a.MemberService, a.ChatService, sessionService, "rests", model.StatusAdmin))))
	a.Dp.AddHandler(handlers.NewCallback(callbackquery.Prefix("reject:"), cf.WrapCallback(restHandler.RejectRestRequest)))
	a.Dp.AddHandler(cf.New(memberHandler.ShowEmoji, "значок").
		WithGuards(groupGuard).
		FallbackToSender().
		WithAmbiguityResolution("show_member_emoji"),
	)

	a.Dp.AddHandler(cf.New(memberHandler.SetEmoji, "значок").
		SetArgsCount(1).
		FallbackToSender().WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin).
		WithDescription("Настройка значка участника").
		WithCategory(cmd.CategoryStats),
	)

	a.Dp.AddHandler(cf.New(memberHandler.RemoveEmoji, "значок").
		AddTriggers("-").
		FallbackToSender().WithGuards(groupGuard).
		Restricted(model.StatusSeniorAdmin),
	)
	a.Dp.AddHandler(handlers.NewMessage(message.LeftChatMember, cf.WrapEvent(memberHandler.OnLeftMember)))
	a.Dp.AddHandler(handlers.NewMyChatMember(chatmember.NewStatus("administrator"), cf.WrapEvent(memberHandler.OnBotPromote)))
	a.Dp.AddHandler(handlers.NewMessage(message.NewChatMembers, cf.WrapEvent(memberHandler.OnJoinMember)))
	a.Dp.AddHandler(handlers.NewMessage(message.Channel, cf.WrapEvent(channelHandler.Post)).SetAllowChannel(true))
	a.Dp.AddHandler(cf.New(channelHandler.Subscribe, "subscribe").WithGuards(groupGuard).Restricted(model.StatusSeniorAdmin).WithDescription("Подписка на канал").WithCategory(cmd.CategoryStats))
	a.Dp.AddHandler(handlers.NewCallback(callbackquery.Prefix("unsubscribe"), cf.WrapEvent(channelHandler.Unsubscribe)))
	a.Dp.AddHandler(handlers.NewMessage(filter.OnlyGroups, cf.WrapEvent(messageHandler.Message)))

}
