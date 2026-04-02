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

	f := command.NewCommandFactory(a.UserService, a.MemberService, a.ChatService, sessionService, a.Config.BotOwnerID, "фм", "!", "/", ".")

	helpHandler := helpH.New(a.Config.BotOwnerUsername, a.Config.CommandsLink)
	statsHandler := statsH.New(statsService, restService, a.MemberService, a.UserService, a.ChatService, sessionService)
	chatHandler := chatH.New(a.ChatService, a.AdminService, a.MemberService, sessionService, dateParser)
	restHandler := restH.New(restService, a.UserService, a.MemberService, a.ChatService, a.AdminService, dateParser, sessionService, a.AsyncClient)

	adminHandler := adminH.New(a.AdminService, a.MemberService, a.ChatService, dateParser, a.AsyncClient, f)

	messageHandler := messageH.New(messageService, a.MemberService, a.ChatService, a.Deepseek)
	memberHandler := memberH.New(a.MemberService, a.ChatService, a.UserService, callService, a.AdminService)
	callHandler := callH.New(callService, a.MemberService, a.ChatService, a.AdminService, sessionService)
	userHandler := userH.New(a.UserService)
	channelHandler := channelH.New(a.MemberService, a.ChatService, a.AsyncClient, a.Config.ChannelID)

	//developerGuard := guard.NewDeveloperGuard(a.AdminService, a.Config.BotOwnerID)
	//groupGuard := guard.OnlyGroups(sessionService)
	//rateLimiterGuard := guard.NewRateLimiter(a.Rdb, 2, 10*time.Second, sessionService)

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
		SetCategory(command.CategoryCall).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.TextRule()),
	)
	a.Dp.AddHandler(f.New("delete_call_message", callHandler.DeleteWelcomeCallMessage).
		SetAliases("калл сообщение").
		AddTriggers("-").
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
		SetCategory(command.CategorySettings),
	)
	a.Dp.AddHandler(f.New("stats", statsHandler.ShowStats).
		SetAliases("отчёт", "отчет", "стата").
		SetArgRules(command.OptionalDateRangeRule()).
		//WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 2, 4*time.Second, sessionService)).
		SetDescription("Отчёт в чате").
		SetCategory(command.CategoryStats),
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
	a.Dp.AddHandler(f.New("unban", adminHandler.Unban).
		SetAliases("unban", "-бан", "разбан", "разбанить").
		SetArgRules(command.MentionedUserRule()).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Разбан участника").
		SetCategory(command.CategoryModeration),
	)
	a.Dp.AddHandler(f.New("unmute", adminHandler.Unmute).
		SetAliases("unmute", "размут", "размутить", "-мут").
		SetArgRules(command.MentionedUserRule()).
		SetRequiredStatus(model.StatusAdmin).
		SetDescription("Размут участника").
		SetCategory(command.CategoryModeration),
	)
	a.Dp.AddHandler(f.New("unwarn", adminHandler.Unwarn).
		SetAliases("снять пред", "-варн", "-пред").
		SetArgRules(command.MentionedUserRule()).
		SetRequiredStatus(model.StatusAdmin).
		SetDescription("Снять предупреждение").
		SetCategory(command.CategoryModeration),
	)

	a.Dp.AddHandler(f.New("kick", adminHandler.Kick).
		SetAliases("кик", "выгнать").
		SetArgRules(command.MentionedUserRule(), command.TextRule().SetVariadic(true).SetRange(0, 1)).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Кик участника").
		SetCategory(command.CategoryModeration),
	)
	a.Dp.AddHandler(f.New("ban", adminHandler.Ban).SetAliases("бан").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.MentionedUserRule(), command.TextRule()).
		SetDescription("Бан участника").
		SetCategory(command.CategoryModeration),
	)
	a.Dp.AddHandler(f.New("mute", adminHandler.Mute).SetAliases("мут").
		SetRequiredStatus(model.StatusAdmin).
		SetArgRules(
			command.MentionedUserRule(),
			command.OneDateRule().SetRange(0, 1),
			command.TextRule().SetVariadic(true).SetRange(0, 1),
		).
		SetDescription("Мут участника").
		SetCategory(command.CategoryModeration),
	)
	a.Dp.AddHandler(f.New("warns", adminHandler.ShowWarns).
		SetArgRules(command.AnyUserRule()).
		SetAliases("варны", "преды"),
	)
	a.Dp.AddHandler(f.New("warnlist", adminHandler.WarnList).SetAliases("варнлист", "предывсе"))
	a.Dp.AddHandler(f.New("warn", adminHandler.Warn).
		SetAliases("варн", "пред").
		SetArgRules(command.MentionedUserRule(), command.TextRule().SetVariadic(true).SetRange(0, 1)).
		SetRequiredStatus(model.StatusAdmin).
		SetDescription("Предупреждение").
		SetCategory(command.CategoryModeration),
	)
	a.Dp.AddHandler(f.New("clear_warns", adminHandler.ClearWarns).SetAliases("очистить преды", "очистить варны").
		SetRequiredStatus(model.StatusAdmin).
		SetArgRules(command.MentionedUserRule()).
		SetDescription("Очистить предупреждения").
		SetCategory(command.CategoryModeration),
	)
	a.Dp.AddHandler(f.New("max_warns", adminHandler.ShowMaxWarns).
		SetAliases("макс преды", "макс варны").
		SetCategory(command.CategoryModeration))
	a.Dp.AddHandler(f.New("set_max_warns", adminHandler.SetMaxWarns).
		SetAliases("max_warns", "макс преды", "макс варны").
		AddTriggers("+").
		SetArgRules(command.NumberRule()).
		SetRequiredStatus(model.StatusCoOwner),
	)
	a.Dp.AddHandler(f.New("rights", adminHandler.ToggleRights).
		SetAliases("права", "rights").
		SetDevCommand(true).
		SetArgRules(command.AnyUserRule(), command.NumberRule()),
	)
	a.Dp.AddHandler(f.New("update_chats", adminHandler.UpdateChats))
	a.Dp.AddHandler(f.New("update_chat", memberHandler.UpdateMembersList).
		SetAliases("обновить чат", "update").
		SetRequiredStatus(model.StatusModerator).
		SetDescription("Обновление списка участников").
		SetCategory(command.CategorySettings),
	)
	a.Dp.AddHandler(f.New("roles", memberHandler.ListRoles).
		SetAliases("роли", "titles").
		SetRequiredStatus(model.StatusMember).
		SetDescription("Список ролей (тегов) участников").
		SetCategory(command.CategoryStats),
	)
	a.Dp.AddHandler(f.New("role", memberHandler.ShowRole).
		SetAliases("роль", "title", "какая роль", "роль у", "роль кого"),
	)
	a.Dp.AddHandler(f.New("set_role", memberHandler.SetRole).
		SetAliases("роль", "title").
		AddTriggers("+").
		SetRequiredStatus(model.StatusModerator).
		SetArgRules(command.AnyUserRule(), command.TextRule().SetVariadic(true)).
		SetDescription("Присвоение ролей участникам").
		SetCategory(command.CategoryStats),
	)
	a.Dp.AddHandler(f.New("move_admins", memberHandler.RestoreRoles).
		SetAliases("перенести админки").
		SetRequiredStatus(model.StatusModerator).
		SetDescription("Перенос тг админок").
		SetCategory(command.CategoryModeration),
	)

	a.Dp.AddHandler(f.New("call_type", callHandler.ShowCallTypes).
		SetAliases("калл тип", "калл стиль"),
	)
	a.Dp.AddHandler(handlers.NewCallback(
		callbackquery.Prefix("call_type:"),
		f.New("call_type", callHandler.CallbackCallType).SetRequiredStatus(model.StatusCoOwner).WrapCallback()),
	)

	a.Dp.AddHandler(f.New("set_call_limit", callHandler.SetMentionsPerMessage).
		SetAliases("калл лимит").
		SetArgRules(command.NumberRule()).
		SetRequiredStatus(model.StatusCoOwner),
	)

	a.Dp.AddHandler(f.New("show_call_limit", callHandler.SetMentionsPerMessage).
		SetAliases("калл лимит"),
	)

	a.Dp.AddHandler(f.New("call_inactive", callHandler.CallInactive).
		SetAliases("калл инактив", "калл неактив", "созвать неактивных").
		SetRequiredStatus(model.StatusModerator).
		SetDescription("Сбор неактивных").
		SetCategory(command.CategoryCall),
	)

	a.Dp.AddHandler(f.New("call_no_norm", callHandler.CallNoNorm).
		SetAliases("калл без нормы", "созвать без нормы").
		SetRequiredStatus(model.StatusModerator).
		SetDescription("Сбор тех, кто без нормы").
		SetCategory(command.CategoryCall),
	)

	a.Dp.AddHandler(f.New("call_no_norm_warn", callHandler.CallNoNormWarn).
		SetAliases("калл без нормы варн", "созвать без нормы варн").
		SetRequiredStatus(model.StatusModerator),
	)

	a.Dp.AddHandler(f.New("call_no_norm_ban", callHandler.CallNoNormBan).
		SetAliases("калл без нормы бан", "созвать без нормы бан").
		SetRequiredStatus(model.StatusModerator),
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

	a.Dp.AddHandler(f.New("call", callHandler.Call).SetAliases("калл", "колл", "all", "каллалл").
		SetRequiredStatus(model.StatusModerator).
		SetArgRules(command.TextRule().SetVariadic(false).SetRange(0, 1)).
		SetDescription("Общий сбор чата").
		SetCategory(command.CategoryCall),
	)

	a.Dp.AddHandler(f.New("prompt", chatHandler.ShowPrompt))

	a.Dp.AddHandler(
		handlers.NewMessage(cmd.NewChatTitle,
			f.New("chat_title_change", chatHandler.OnNewChatTitle).WrapEvent()),
	)
	a.Dp.AddHandler(f.New("manage", chatHandler.Manage).
		SetAliases("управление").
		SetScope(command.ScopeUser),
	)
	a.Dp.AddHandler(
		handlers.NewCallback(callbackquery.Prefix("manage:"),
			f.New("set_manage", chatHandler.CallbackManage).WrapCallback()),
	)

	a.Dp.AddHandler(
		handlers.NewCallback(callbackquery.Prefix("manage_page:"),
			f.New("manage_page", chatHandler.CallbackManagePage).WrapCallback()),
	)

	a.Dp.AddHandler(f.New("enable_tags", chatHandler.EnableTags).
		SetAliases("+tags", "+теги", "+тэги").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Включение тегов").
		SetCategory(command.CategorySettings),
	)
	a.Dp.AddHandler(f.New("disable_tags", chatHandler.DisableTags).SetAliases("-tags", "-теги", "-тэги").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Отключение тегов").
		SetCategory(command.CategorySettings),
	)
	a.Dp.AddHandler(f.New("tags", chatHandler.ShowTags).
		SetAliases("tags", "теги", "тэги"))

	a.Dp.AddHandler(f.New("chats", chatHandler.UserChats).
		SetAliases("чаты", "нормы", "чаты без нормы"))

	a.Dp.AddHandler(f.New("set_prompt", chatHandler.SetPrompt).SetAliases("промпт").
		AddTriggers("+").
		SetArgRules(command.TextRule().SetVariadic(false).SetRange(0, 1)).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Настройка промпта для ИИ").
		SetCategory(command.CategorySettings),
	)
	a.Dp.AddHandler(f.New("week_start", chatHandler.ShowWeekStart).
		SetAliases("начало недели", "чистка", "время чистки", "конец чистки"))

	a.Dp.AddHandler(f.New("set_week_start", chatHandler.SetWeekStart).
		SetAliases("начало недели", "чистка", "время чистки", "конец чистки").
		AddTriggers("+").
		SetArgRules(command.OneDateRule()).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Настройка начала недели").
		SetCategory(command.CategorySettings),
	)
	a.Dp.AddHandler(f.New("custom_prefix", chatHandler.ShowPrefix).
		SetAliases("кастом префикс", "префикс").
		SetRequiredStatus(model.StatusSeniorAdmin),
	)
	a.Dp.AddHandler(f.New("set_custom_prefix", chatHandler.SetPrefix).
		SetAliases("кастом префикс", "префикс").
		AddTriggers("+").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.TextRule()).
		SetDescription("Кастомные префиксы").
		SetCategory(command.CategorySettings),
	)
	a.Dp.AddHandler(f.New("prefix_only", chatHandler.ShowPrefixlessStatus).
		SetAliases("префиксы", "prefixes"),
	)
	a.Dp.AddHandler(f.New("enable_prefix_only", chatHandler.EnablePrefixOnly).
		SetAliases("с префикса", "с префиксами", "-prefixless", "+префиксы").
		SetRequiredStatus(model.StatusSeniorAdmin),
	)

	a.Dp.AddHandler(f.New("disable_prefix_only", chatHandler.DisablePrefixOnly).
		SetAliases("без префикса", "без префиксов", "+prefixless", "-префиксы").
		SetRequiredStatus(model.StatusSeniorAdmin),
	)
	a.Dp.AddHandler(f.New("enable_call_on_join", callHandler.EnableCallOnJoin).
		SetAliases("call_enable", "включить call", "включить колл", "включить калл").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Включить призыв при входе").
		SetCategory(command.CategoryCall),
	)
	a.Dp.AddHandler(f.New("call_disable", callHandler.DisableCallOnJoin).
		SetAliases("отключить call", "отключить колл", "отключить калл").
		SetRequiredStatus(model.StatusSeniorAdmin),
	)
	a.Dp.AddHandler(f.New("ask_ai", messageHandler.Bot).SetAliases("крис").
		SetArgRules(command.TextRule().SetVariadic(true).SetRange(0, 1)),
	)

	a.Dp.AddHandler(f.New("demote", adminHandler.DemoteTgAdmin).
		SetAliases("разжаловать").
		SetArgRules(command.MentionedUserRule()).
		SetRequiredStatus(model.StatusSeniorAdmin))

	a.Dp.AddHandler(f.New("fake_leave", adminHandler.FakeLeave).
		SetAliases("фейклив", "фейк лив").SetArgRules(command.AnyUserRule()))

	a.Dp.AddHandler(f.New("set_gender", userHandler.SetGender).
		SetAliases("пол", "установить пол").
		SetArgRules(command.TextRule()),
	)
	a.Dp.AddHandler(f.New("gender", userHandler.ShowGender).
		SetAliases("пол").
		SetArgRules(command.AnyUserRule()))
	a.Dp.AddHandler(f.New("set_emoji", userHandler.SetEmoji).
		SetAliases("эмоджи", "эмодзи").
		SetArgRules(command.AnyUserRule(), command.TextRule().SetVariadic(true).SetRange(0, 1)))

	a.Dp.AddHandler(f.New("emoji", userHandler.ShowEmoji).SetAliases("эмоджи", "эмодзи").
		SetArgRules(command.AnyUserRule()),
	)
	a.Dp.AddHandler(f.New("remove_emoji", userHandler.RemoveEmoji).
		SetAliases("-эмоджи", "-эмодзи").
		SetArgRules(command.AnyUserRule()))

	a.Dp.AddHandler(f.New("manage_rights", adminHandler.ManageRights).
		SetAliases("дк").
		SetRequiredStatus(model.StatusCoOwner).
		SetDescription("Управление доступом команд").
		SetCategory(command.CategorySettings),
	)
	a.Dp.AddHandler(
		handlers.NewCallback(callbackquery.Prefix("rights_"),
			f.New("manage_rights_callback", adminHandler.CallbackManageRights).SetRequiredStatus(model.StatusCoOwner).WrapCallback()))

	a.Dp.AddHandler(
		handlers.NewCallback(callbackquery.Prefix("approve:"),
			f.New("approve_rest_request", restHandler.ApproveRestRequest).WrapCallback()))

	a.Dp.AddHandler(
		handlers.NewCallback(callbackquery.Prefix("reject:"),
			f.New("reject_rest_request", restHandler.RejectRestRequest).WrapCallback()))

	a.Dp.AddHandler(f.New("set_member_emoji", memberHandler.SetEmoji).SetAliases("значок").
		SetArgRules(command.AnyUserRule(), command.TextRule()).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Настройка значка участника").
		SetCategory(command.CategoryStats),
	)
	a.Dp.AddHandler(f.New("member_emoji", memberHandler.ShowEmoji).
		SetAliases("значок").
		SetArgRules(command.AnyUserRule()),
	)

	a.Dp.AddHandler(f.New("remove_member_emoji", memberHandler.RemoveEmoji).
		SetAliases("значок").
		AddTriggers("-").
		SetArgRules(command.AnyUserRule()).
		SetRequiredStatus(model.StatusSeniorAdmin),
	)
	a.Dp.AddHandler(
		handlers.NewMessage(message.LeftChatMember,
			f.New("left_member", memberHandler.OnLeftMember).WrapEvent()))
	a.Dp.AddHandler(
		handlers.NewMyChatMember(chatmember.NewStatus("administrator"),
			f.New("bot_promote", memberHandler.OnBotPromote).WrapEvent()))
	a.Dp.AddHandler(
		handlers.NewMessage(message.NewChatMembers,
			f.New("new_members", memberHandler.OnJoinMember).WrapEvent()))
	a.Dp.AddHandler(
		f.New("subscribe", channelHandler.Subscribe).
			SetRequiredStatus(model.StatusSeniorAdmin).
			SetDescription("Подписка на канал").
			SetCategory(command.CategoryStats))
	a.Dp.AddHandler(
		handlers.NewCallback(callbackquery.Prefix("unsubscribe"),
			f.New("unsubscribe", channelHandler.Unsubscribe).WrapCallback()))
	a.Dp.AddHandler(
		handlers.NewMessage(filter.OnlyGroups, f.New("message", messageHandler.Message).WrapEvent()))

}
