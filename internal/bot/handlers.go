package bot

import (
	"activity-bot/internal/call"
	callH "activity-bot/internal/call/handler"
	chatH "activity-bot/internal/chat/handler"
	"activity-bot/internal/command"
	"activity-bot/internal/conversation"
	"activity-bot/internal/db/postgres"
	db "activity-bot/internal/db/postgres/sqlc"
	helpH "activity-bot/internal/help/handler"
	"activity-bot/internal/helpers"
	"activity-bot/internal/logger"
	"activity-bot/internal/middleware"
	"activity-bot/internal/model"
	"activity-bot/internal/rest"
	restH "activity-bot/internal/rest/handler"
	"activity-bot/internal/session"
	"activity-bot/internal/stats"
	statsH "activity-bot/internal/stats/handler"
	"context"
	time "time"

	adminH "activity-bot/internal/admin/handler"
	channelH "activity-bot/internal/channel/handler"
	memberH "activity-bot/internal/member/handler"
	msg "activity-bot/internal/message"
	messageH "activity-bot/internal/message/handler"
	userH "activity-bot/internal/user/handler"

	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/gotd/td/tg"
)

func (a *App) RegisterHandlers() {
	queries := db.New(a.Pool)

	statsRepository := postgres.NewStatsRepository(queries)
	restRepository := postgres.NewRestRepository(queries, a.Pool)
	messageRepository := postgres.NewMessageRepository(queries)
	sessionRepository := postgres.NewSessionRepository(queries)

	statsService := stats.NewService(statsRepository)
	restService := rest.NewService(restRepository)
	messageService := msg.NewService(messageRepository)

	callService := call.NewService(postgres.NewChatRepository(queries), a.MemberService, statsService)
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
	storage := conversation.NewRedisStorage(a.Rdb, "convo")
	callHandler := callH.New(callService, a.MemberService, a.ChatService, a.AdminService, sessionService, storage)
	userHandler := userH.New(a.UserService)
	channelHandler := channelH.New(a.MemberService, a.ChatService, a.AsyncClient, a.Config.ChannelID)

	rateLimiterMiddleware := middleware.NewRateLimiter(a.Rdb, 3, 10*time.Second)

	a.dp.AddHandler(f.New("start", helpHandler.Start).SetScope(command.ScopeUser))
	a.dp.AddHandler(f.New("help", helpHandler.Help).SetScope(command.ScopeUser))

	a.dp.AddHandler(f.New("set_newbie_treshold", chatHandler.SetNewbieThreshold).
		SetAliases("новички срок", "новички после").
		AddPrefixes("+").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.NumberRule()).
		SetDescription("Настройка срока новичка").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("show_newbie_treshold", chatHandler.ShowNewbieThreshold).
		SetDescription("Порог новичка").
		SetCategory(command.CategorySettings).
		SetAliases("новички срок", "новички после").
		SetRequiredStatus(model.StatusSeniorAdmin),
	)
	a.dp.AddHandler(f.New("norm", chatHandler.ShowNorm).
		SetDescription("Посмотреть норму сообщений").
		SetCategory(command.CategorySettings).
		SetAliases("норма какая", "а норма какая", "норма", "норма?", "quota", "какая норма", "а какая норма"),
	)
	a.dp.AddHandler(f.New("set_norm", chatHandler.SetNorm).
		SetDescription("Установка нормы").
		SetCategory(command.CategorySettings).
		SetAliases("норма", "quota").
		AddPrefixes("+").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.NumberRule(), command.TextRule().SetVariadic(true).SetRange(0, 1)),
	)
	a.dp.AddHandler(f.New("remove_norm", chatHandler.RemoveNorm).
		SetDescription("Удалить норму").
		SetCategory(command.CategorySettings).
		SetAliases("норма", "quota").
		AddPrefixes("-").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.TextRule().SetVariadic(true).SetRange(0, 1)),
	)

	a.dp.AddHandler(f.New("prompt", chatHandler.ShowPrompt).
		SetDescription("Системный ИИ промпт").
		SetCategory(command.CategorySettings))
	a.dp.AddHandler(
		f.New("new_members", memberHandler.OnJoinMember).WrapEvent(joinMemberFilter),
	)
	a.dp.AddHandler(f.New("chat_title_change", chatHandler.OnNewChatTitle).WrapEvent(chatTitleChangedFilter))

	a.dp.AddHandler(f.New("manage", chatHandler.Manage).
		SetDescription("Управление чатом").
		SetCategory(command.CategorySettings).
		SetAliases("управление").
		SetScope(command.ScopeUser),
	)

	a.dp.AddHandler(
		f.New("set_manage", chatHandler.CallbackManage).
			SetScope(command.ScopeUser).
			WrapCallback(filters.CallbackQuery.Prefix("manage:")),
	)

	a.dp.AddHandler(
		f.New("manage_page", chatHandler.CallbackManagePage).
			WrapCallback(filters.CallbackQuery.Prefix("manage_page:")),
	)

	a.dp.AddHandler(f.New("enable_tags", chatHandler.EnableTags).
		SetAliases("+tags", "+теги", "+тэги").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Включение тегов").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("disable_tags", chatHandler.DisableTags).SetAliases("-tags", "-теги", "-тэги").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Отключение тегов").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("tags", chatHandler.ShowTags).SetDescription("Статус тегов").
		SetCategory(command.CategoryGeneral).
		SetAliases("tags", "теги", "тэги"))

	a.dp.AddHandler(f.New("my_norms", chatHandler.UserChats).SetDescription("Список норм во всех чатах").SetCategory(command.CategoryAdmin).
		SetAliases("чаты", "нормы", "чаты без нормы"))

	a.dp.AddHandler(f.New("set_prompt", chatHandler.SetPrompt).SetAliases("промпт").
		AddPrefixes("+").
		SetArgRules(command.TextRule().SetVariadic(true)).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Настройка промпта для ИИ").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("week_start", chatHandler.ShowWeekStart).SetDescription("Начало недели для нормы").SetCategory(command.CategorySettings).
		SetAliases("начало недели", "чистка", "время чистки", "конец чистки"))

	a.dp.AddHandler(f.New("set_week_start", chatHandler.SetWeekStart).
		SetAliases("начало недели", "чистка", "время чистки", "конец чистки").
		AddPrefixes("+").
		SetArgRules(command.OneDateRule()).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Настройка начала недели").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("custom_prefix", chatHandler.ShowPrefix).
		SetAliases("кастом префикс", "префикс").
		SetRequiredStatus(model.StatusSeniorAdmin),
	)
	a.dp.AddHandler(f.New("set_custom_prefix", chatHandler.SetPrefix).
		SetAliases("кастом префикс", "префикс").
		AddPrefixes("+").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.TextRule()).
		SetDescription("Кастомные префиксы").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("show_newbie_treshold", chatHandler.ShowNewbieThreshold).SetDescription("Порог новичка").SetCategory(command.CategorySettings).
		SetAliases("новички срок").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Показать срок новичка").
		SetCategory(command.CategorySettings),
	)

	a.dp.AddHandler(f.New("remove_norm", chatHandler.RemoveNorm).SetDescription("Удалить норму").SetCategory(command.CategorySettings).
		SetAliases("норма").
		AddPrefixes("-").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Удалить норму").
		SetCategory(command.CategorySettings),
	)

	a.dp.AddHandler(f.New("prompt", chatHandler.ShowPrompt).
		SetDescription("Системный ИИ промпт").
		SetCategory(command.CategorySettings).
		SetAliases("промпт").
		SetDescription("Показать текущий AI-промпт").
		SetCategory(command.CategorySettings),
	)

	a.dp.AddHandler(f.New("week_start", chatHandler.ShowWeekStart).SetDescription("Начало недели для нормы").SetCategory(command.CategorySettings).
		SetAliases("начало недели").
		SetDescription("Показать начало недели").
		SetCategory(command.CategorySettings),
	)

	a.dp.AddHandler(f.New("custom_prefix", chatHandler.ShowPrefix).
		SetAliases("префикс").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Показать префиксы команд").
		SetCategory(command.CategorySettings),
	)

	a.dp.AddHandler(f.New("prefix_only", chatHandler.ShowPrefixlessStatus).
		SetAliases("префиксы").
		SetDescription("Статус обязательного префикса").
		SetCategory(command.CategorySettings),
	)

	a.dp.AddHandler(f.New("enable_prefix_only", chatHandler.EnablePrefixOnly).
		SetAliases("+префиксы", "с префиксами").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Включить обязательный префикс").
		SetCategory(command.CategorySettings),
	)

	a.dp.AddHandler(f.New("disable_prefix_only", chatHandler.DisablePrefixOnly).
		SetAliases("-префиксы").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Отключить обязательный префикс").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("enable_call_on_join", callHandler.EnableCallOnJoin).
		SetAliases("call_enable", "включить call", "включить колл", "включить калл").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Включить призыв при входе").
		SetCategory(command.CategoryCall),
	)
	a.dp.AddHandler(f.New("set_call_message", callHandler.SetWelcomeCallMessage).
		SetAliases("калл сообщение").
		AddPrefixes("+").
		SetDescription("Настройка сообщения сбора").
		SetCategory(command.CategoryCall).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.TextRule()),
	)
	a.dp.AddHandler(f.New("show_call_message", callHandler.ShowWelcomeCallMessage).
		SetAliases("калл сообщение").
		SetProviders(a.UserService, a.MemberService, a.ChatService, sessionService).
		SetDescription("Показать сообщение сбора").
		SetCategory(command.CategoryCall),
	)

	a.dp.AddHandler(f.New("call_disable", callHandler.DisableCallOnJoin).
		SetAliases("отключить калл").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Отключить call при входе").
		SetCategory(command.CategoryCall),
	)
	a.dp.AddHandler(f.New("delete_call_message", callHandler.DeleteWelcomeCallMessage).
		SetAliases("калл сообщение", "удалить калл сообщение").
		AddPrefixes("-").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Удалить сообщение сбора").
		SetCategory(command.CategoryCall),
	)
	a.dp.AddHandler(f.New("call_type", callHandler.ShowCallTypes).
		SetAliases("калл тип").
		SetDescription("Показать типы сбора").
		SetCategory(command.CategoryCall),
	)
	a.dp.AddHandler(
		f.New("call_type_callback", callHandler.ShowCallTypes).
			WrapCallback(filters.CallbackQuery.Equal("call_type")),
	)
	a.dp.AddHandler(
		f.New("call_type", callHandler.CallbackCallType).
			SetRequiredStatus(model.StatusCoOwner).
			WrapCallback(filters.CallbackQuery.Prefix("call_type:")),
	)

	a.dp.AddHandler(f.New("call_no_norm_warn", callHandler.CallNoNormWarn).
		SetAliases("калл без нормы варн").
		SetRequiredStatus(model.StatusModerator).
		SetDescription("Сбор без нормы с предупреждением").
		SetCategory(command.CategoryCall),
	)
	a.dp.AddHandler(f.New("call_no_norm", callHandler.CallNoNorm).
		SetAliases("калл без нормы", "созвать без нормы").
		SetRequiredStatus(model.StatusModerator).
		SetDescription("Сбор тех, кто без нормы").
		SetCategory(command.CategoryCall),
	)

	a.dp.AddHandler(f.New("call_no_norm_warn", callHandler.CallNoNormWarn).
		SetAliases("калл без нормы варн", "созвать без нормы варн").
		SetRequiredStatus(model.StatusModerator),
	)
	a.dp.AddHandler(f.New("call_no_norm_ban", callHandler.CallNoNormBan).
		SetAliases("калл без нормы бан").
		SetRequiredStatus(model.StatusModerator).
		SetDescription("Сбор без нормы с баном").
		SetCategory(command.CategoryCall),
	)
	a.dp.AddHandler(f.New("set_call_limit", callHandler.SetMentionsPerMessage).
		SetAliases("калл лимит").
		SetArgRules(command.NumberRule()).
		SetRequiredStatus(model.StatusCoOwner).
		SetDescription("Лимит упоминаний в call").
		SetCategory(command.CategoryCall),
	)
	a.dp.AddHandler(f.New("show_call_limit", callHandler.SetMentionsPerMessage).
		SetAliases("калл лимит").
		SetDescription("Показать лимит call").
		SetCategory(command.CategoryCall),
	)

	a.dp.AddHandler(f.New("call_inactive", callHandler.CallInactive).
		SetAliases("калл инактив", "калл неактив", "созвать неактивных").
		SetRequiredStatus(model.StatusModerator).
		SetArgRules(command.TextRule().SetVariadic(true).SetRange(0, 1)).
		SetDescription("Сбор неактивных").
		SetCategory(command.CategoryCall),
	)

	callConversation := conversation.NewConversation(
		[]conversation.Handler{
			f.New("start_call_inactive", callHandler.StartCallInactiveConversation).
				SetRequiredStatus(model.StatusModerator).
				WrapCallback(filters.CallbackQuery.Equal("call_inactive")),
			f.New("start_call_no_norm_warn", callHandler.StartCallNoNormWarnConversation).
				SetRequiredStatus(model.StatusModerator).
				WrapCallback(filters.CallbackQuery.Equal("call_no_norm_warn")),
			f.New("start_call_no_norm_ban", callHandler.StartCallNoNormBanConversation).
				SetRequiredStatus(model.StatusModerator).
				WrapCallback(filters.CallbackQuery.Equal("call_no_norm_ban")),
			f.New("start_call_no_norm", callHandler.StartCallNoNormConversation).
				SetRequiredStatus(model.StatusModerator).
				WrapCallback(filters.CallbackQuery.Equal("call_no_norm")),
		},
		map[string][]conversation.Handler{
			callH.CallStateInactive: {
				f.New("handle_call_inactive_message", callHandler.HandleCallInactiveMessage).WrapEvent(textMessageFilter),
			},
			callH.CallStateNoNorm: {
				f.New("handle_call_no_norm_message", callHandler.HandleCallNoNormMessage).WrapEvent(textMessageFilter),
			},
			callH.CallStateNoNormWarn: {
				f.New("handle_call_no_norm_warn_message", callHandler.HandleCallNoNormWarnMessage).WrapEvent(textMessageFilter),
			},
			callH.CallStateNoNormBan: {
				f.New("handle_call_no_norm_ban_message", callHandler.HandleCallNoNormBanMessage).WrapEvent(textMessageFilter),
			},
		},
		storage,
		conversation.WithExits(
			f.New("no_msg_call_convo", callHandler.NoMessageCallConversation).WrapCallback(filters.CallbackQuery.Prefix("call_nomsg:")),
			f.New("cancel_call_convo", callHandler.CancelCallConversation).WrapCallback(filters.CallbackQuery.Equal("call_cancel")),
		),
	)
	a.dp.AddHandler(callConversation)
	a.dp.AddHandler(f.New("call", callHandler.Call).
		SetAliases("калл", "колл", "all", "каллалл").
		SetRequiredStatus(model.StatusModerator).
		SetArgRules(command.TextRule().SetVariadic(true).SetRange(0, 1)).
		SetDescription("Общий сбор чата").
		SetCategory(command.CategoryCall),
	)
	a.dp.AddHandler(f.New("stats", statsHandler.ShowStats).
		SetAliases("отчёт", "отчет", "стата").
		SetArgRules(command.OptionalDateRangeRule()).
		//WithGuards(groupGuard, middleware.NewRateLimiter(a.Rdb, 2, 4*time.Second, sessionService)).
		SetDescription("Отчёт в чате").
		SetCategory(command.CategoryStats),
	)
	a.dp.AddHandler(f.New("stats_graph", statsHandler.ShowChatActivityGraph).SetDescription("График активности чата").SetCategory(command.CategoryStats).
		SetAliases("график").
		SetArgRules(command.OptionalDateRangeRule()),
	//WithGuards(groupGuard, rateLimiterGuard),
	)
	a.dp.AddHandler(f.New("rests", statsHandler.ShowRestList).
		SetDescription("Список участников в ресте").
		SetAliases("ресты"),
	)

	a.dp.AddHandler(f.New("who_am_i", statsHandler.WhoAmI).
		SetDescription("Мой профиль").
		SetCategory(command.CategoryProfile).
		SetAliases("ктоя", "кто я", "профиль"),
	)
	a.dp.AddHandler(f.New("callback_all_activity", statsHandler.CallbackAllActivity).
		WrapCallback(filters.CallbackQuery.Prefix("profile_activity:")),
	)
	a.dp.AddHandler(f.New("callback_profile_graph", statsHandler.CallbackProfileGraph).
		WrapCallback(filters.CallbackQuery.Prefix("profile_graph:")),
	)
	a.dp.AddHandler(f.New("ship_random", memberHandler.ShipRandom).
		SetDescription("Случайный шипперинг").
		SetCategory(command.CategoryGeneral).
		SetAliases("рандом шипперим", "шипперим"))

	a.dp.AddHandler(f.New("rests", restHandler.ShowRest).
		SetDescription("История рестов").
		SetCategory(command.CategoryProfile).
		SetAliases("все ресты").SetArgRules(command.AnyUserRule()),
	)

	a.dp.AddHandler(f.New("all_rests", restHandler.AllUserRests).
		SetDescription("История рестов").
		SetCategory(command.CategoryProfile).
		SetAliases("все ресты").SetArgRules(command.AnyUserRule()),
	)

	a.dp.AddHandler(f.New("set_rest", restHandler.SetRest).SetDescription("Выдать рест").
		SetCategory(command.CategoryModeration).
		SetRequiredStatus(model.StatusModerator).
		DisableCheckStatus().
		SetAliases("рест", "rest", "установить рест").
		AddPrefixes("+").
		SetArgRules(command.AnyUserRule(), command.OneDateRule()),
	)
	a.dp.AddHandler(f.New("rest", restHandler.ShowRest).SetDescription("Информация о ресте").SetCategory(command.CategoryProfile).
		SetAliases("рест", "rest").
		SetArgRules(command.AnyUserRule()),
	)
	a.dp.AddHandler(f.New("remove_rest", restHandler.RemoveRestRequest).SetDescription("Удалить рест").SetCategory(command.CategoryModeration).
		SetAliases("удалить рест").
		SetRequiredStatus(model.StatusAdmin).
		SetArgRules(command.AnyUserRule(), command.NumberRule()),
	)
	a.dp.AddHandler(f.New("end_rest", restHandler.EndRest).SetDescription("Досрочно снять рест").SetCategory(command.CategoryModeration).SetAliases("рест", "rest", "снять рест").
		AddPrefixes("-").
		SetRequiredStatus(model.StatusModerator).
		SetArgRules(command.AnyUserRule()),
	)
	a.dp.AddHandler(
		f.New("approve_rest_request", restHandler.ApproveRestRequest).
			WrapCallback(filters.CallbackQuery.Prefix("approve:")),
	)
	a.dp.AddHandler(
		f.New("reject_rest_request", restHandler.RejectRestRequest).
			WrapCallback(filters.CallbackQuery.Prefix("reject:")),
	)
	a.dp.AddHandler(f.New("admins", adminHandler.ListAdmins).SetDescription("Список администрации").SetCategory(command.CategoryGeneral).
		SetAliases("админы", "модеры", "mods"),
	)

	a.dp.AddHandler(f.New("add_admin", adminHandler.SetStatus).
		SetDescription("Назначить администратора").
		SetCategory(command.CategoryAdmin).
		SetAliases("админ", "admin", "mod", "повысить").
		SetArgRules(command.AnyUserRule(), command.NumberRule()).AddPrefixes("+").
		SetRequiredStatus(model.StatusCoOwner),
	)
	a.dp.AddHandler(f.New("is_admin", adminHandler.IsAdmin).SetDescription("Проверка прав администратора").SetCategory(command.CategoryAdmin).SetAliases("админ", "admin", "mod").
		SetArgRules(command.AnyUserRule()),
	)
	a.dp.AddHandler(f.New("remove_admin", adminHandler.RemoveAdmin).SetDescription("Снять администратора").SetCategory(command.CategoryAdmin).SetAliases("админ", "admin", "mod").
		SetPrefixes("-").SetArgRules(command.MentionedUserRule()).
		SetRequiredStatus(model.StatusCoOwner),
	)
	a.dp.AddHandler(f.New("unban", adminHandler.Unban).
		SetAliases("unban", "-бан", "разбан", "разбанить").
		SetArgRules(command.MentionedUserRule()).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Разбан участника").
		SetCategory(command.CategoryModeration),
	)
	a.dp.AddHandler(f.New("unmute", adminHandler.Unmute).
		SetAliases("unmute", "размут", "размутить", "-мут", "снять мут").
		SetArgRules(command.MentionedUserRule()).
		SetRequiredStatus(model.StatusAdmin).
		SetDescription("Размут участника").
		SetCategory(command.CategoryModeration),
	)
	a.dp.AddHandler(f.New("unwarn", adminHandler.Unwarn).
		SetAliases("снять пред", "-варн", "-пред").
		SetArgRules(command.MentionedUserRule()).
		SetRequiredStatus(model.StatusAdmin).
		SetDescription("Снять предупреждение").
		SetCategory(command.CategoryModeration),
	)

	a.dp.AddHandler(f.New("kick", adminHandler.Kick).
		SetAliases("кик", "выгнать").
		SetArgRules(command.MentionedUserRule(), command.TextRule().SetVariadic(true).SetRange(0, 1)).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Кик участника").
		SetCategory(command.CategoryModeration),
	)
	a.dp.AddHandler(f.New("ban", adminHandler.Ban).SetAliases("бан").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.MentionedUserRule(), command.TextRule()).
		SetDescription("Бан участника").
		SetCategory(command.CategoryModeration),
	)
	a.dp.AddHandler(f.New("mute", adminHandler.Mute).
		SetAliases("мут").
		SetRequiredStatus(model.StatusAdmin).
		SetArgRules(
			command.MentionedUserRule(),
			command.OneDateRule().SetRange(0, 1),
			command.TextRule().SetVariadic(true).SetRange(0, 1),
		).
		SetDescription("Мут участника").
		SetCategory(command.CategoryModeration),
	)
	a.dp.AddHandler(f.New("warns", adminHandler.ShowWarns).SetDescription("Список предупреждений").SetCategory(command.CategoryProfile).
		SetArgRules(command.AnyUserRule()).
		SetAliases("варны", "преды"),
	)
	a.dp.AddHandler(f.New("warnlist", adminHandler.WarnList).SetDescription("Предупреждения в чате").SetCategory(command.CategoryModeration).SetAliases("варнлист", "предывсе"))
	a.dp.AddHandler(f.New("warn", adminHandler.Warn).
		SetAliases("варн", "пред").
		SetArgRules(command.MentionedUserRule(), command.TextRule().SetVariadic(true).SetRange(0, 1)).
		SetRequiredStatus(model.StatusAdmin).
		SetDescription("Предупреждение").
		SetCategory(command.CategoryModeration),
	)
	a.dp.AddHandler(f.New("clear_warns", adminHandler.ClearWarns).SetAliases("очистить преды", "очистить варны").
		SetRequiredStatus(model.StatusAdmin).
		SetArgRules(command.MentionedUserRule()).
		SetDescription("Очистить предупреждения").
		SetCategory(command.CategoryModeration),
	)
	a.dp.AddHandler(f.New("max_warns", adminHandler.ShowMaxWarns).SetDescription("Максимальное количество предупреждений").SetCategory(command.CategorySettings).
		SetAliases("макс преды", "макс варны").
		SetCategory(command.CategoryModeration))
	a.dp.AddHandler(f.New("set_max_warns", adminHandler.SetMaxWarns).SetDescription("Установить лимит предупреждений").SetCategory(command.CategorySettings).
		SetAliases("max_warns", "макс преды", "макс варны").
		AddPrefixes("+").
		SetArgRules(command.NumberRule()).
		SetRequiredStatus(model.StatusCoOwner),
	)
	a.dp.AddHandler(f.New("rights", adminHandler.ToggleRights).SetDescription("Управление правами").SetCategory(command.CategoryAdmin).
		SetAliases("права", "rights").
		SetDevCommand(true).
		SetArgRules(command.AnyUserRule(), command.NumberRule()),
	)
	a.dp.AddHandler(f.New("update_chats", adminHandler.UpdateChats).
		SetDescription("Обновить кэш чатов").
		SetCategory(command.CategoryAdmin).
		SetDevCommand(true))
	a.dp.AddHandler(f.New("update_chat", memberHandler.UpdateMembersList).SetDescription("Обновление списка участников").SetCategory(command.CategorySettings).
		SetAliases("обновить чат", "update").
		WithMiddlewares(rateLimiterMiddleware).
		SetRequiredStatus(model.StatusMember).
		SetDescription("Обновление списка участников").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("roles", memberHandler.ListRoles).
		SetAliases("роли", "titles").
		SetRequiredStatus(model.StatusMember).
		SetDescription("Список ролей (тегов) участников").
		SetCategory(command.CategoryStats),
	)
	a.dp.AddHandler(f.New("set_role", memberHandler.SetRole).
		SetAliases("роль", "title").
		AddPrefixes("+").
		SetRequiredStatus(model.StatusModerator).
		SetArgRules(command.AnyUserRule(), command.TextRule().SetVariadic(true)).
		SetDescription("Присвоение ролей участникам").
		SetCategory(command.CategoryStats),
	)
	a.dp.AddHandler(f.New("role", memberHandler.ShowRole).SetDescription("Посмотреть роль").
		SetArgRules(command.AnyUserRule()).
		SetCategory(command.CategoryProfile).
		SetAliases("роль", "title", "какая роль", "роль у", "роль кого"),
	)

	a.dp.AddHandler(f.New("ask_ai", messageHandler.Bot).SetDescription("Вопрос к ИИ").SetCategory(command.CategoryGeneral).SetAliases("крис").
		SetArgRules(command.TextRule().SetVariadic(true).SetRange(0, 1)),
	)

	a.dp.AddHandler(f.New("demote", adminHandler.DemoteTgAdmin).SetDescription("Разжаловать администратора в Telegram").SetCategory(command.CategoryAdmin).
		SetAliases("разжаловать").
		SetArgRules(command.MentionedUserRule()).
		SetRequiredStatus(model.StatusSeniorAdmin))

	a.dp.AddHandler(f.New("fake_leave", adminHandler.FakeLeave).SetDescription("Сымитировать выход из чата").SetCategory(command.CategoryGeneral).
		SetAliases("фейклив", "фейк лив").SetArgRules(command.AnyUserRule()))

	a.dp.AddHandler(f.New("set_gender", userHandler.SetGender).SetDescription("Установить пол").SetCategory(command.CategoryProfile).
		SetAliases("мой пол", "установить пол").SetScope(command.ScopeUser).
		SetArgRules(command.TextRule()),
	)
	a.dp.AddHandler(f.New("gender", userHandler.ShowGender).SetDescription("Посмотреть пол").SetCategory(command.CategoryProfile).
		SetAliases("мой пол").SetScope(command.ScopeUser).
		SetArgRules(command.AnyUserRule()))
	a.dp.AddHandler(f.New("set_emoji", userHandler.SetEmoji).SetDescription("Установить эмодзи").SetCategory(command.CategoryProfile).
		SetAliases("эмоджи", "эмодзи").
		SetArgRules(command.AnyUserRule(), command.TextRule().SetVariadic(true).SetRange(1, 1)))

	a.dp.AddHandler(f.New("emoji", userHandler.ShowEmoji).
		SetDescription("Посмотреть эмодзи").
		SetCategory(command.CategoryProfile).
		SetAliases("эмоджи", "эмодзи").
		SetArgRules(command.AnyUserRule()),
	)
	a.dp.AddHandler(f.New("remove_emoji", userHandler.RemoveEmoji).SetDescription("Удалить эмодзи").SetCategory(command.CategoryProfile).
		SetAliases("-эмоджи", "-эмодзи").
		SetArgRules(command.AnyUserRule()))

	a.dp.AddHandler(f.New("manage_rights", adminHandler.ManageRights).
		SetAliases("дк").
		SetRequiredStatus(model.StatusCoOwner).
		SetDescription("Управление доступом команд").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(
		f.New("manage_rights_callback", adminHandler.CallbackManageRights).
			SetRequiredStatus(model.StatusCoOwner).
			WrapCallback(filters.CallbackQuery.Prefix("rights_")),
	)

	a.dp.AddHandler(f.New("set_member_emoji", memberHandler.SetEmoji).SetAliases("значок").
		SetArgRules(command.AnyUserRule(), command.TextRule()).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Настройка значка участника").
		SetCategory(command.CategoryStats),
	)
	a.dp.AddHandler(f.New("member_emoji", memberHandler.ShowEmoji).SetDescription("Посмотреть значок участника").SetCategory(command.CategoryProfile).
		SetAliases("значок").
		SetArgRules(command.AnyUserRule()),
	)

	a.dp.AddHandler(f.New("remove_member_emoji", memberHandler.RemoveEmoji).
		SetAliases("значок").
		AddPrefixes("-").
		SetArgRules(command.AnyUserRule()).
		SetRequiredStatus(model.StatusSeniorAdmin),
	)
	a.dp.AddHandler(
		f.New("subscribe", channelHandler.Subscribe).SetDescription("Подписка на канал").SetCategory(command.CategorySettings).
			SetRequiredStatus(model.StatusSeniorAdmin),
	)
	a.dp.AddHandler(
		f.New("unsubscribe", channelHandler.Unsubscribe).SetDescription("Отписка от канала").SetCategory(command.CategorySettings).
			WrapCallback(filters.CallbackQuery.Prefix("unsubscribe")),
	)
	a.dp.AddHandler(
		f.New("left_member", memberHandler.OnLeftMember).WrapEvent(leftMemberFilter),
	)

	a.dp.AddHandler(
		f.New("message", messageHandler.Message).WrapEvent(textMessageFilter),
	)

	var userScopeBotCommands []tg.BotCommand
	var chatScopeBotCommands []tg.BotCommand

	for _, cmd := range f.ConfigurableCommands() {
		bc := tg.BotCommand{
			Command:     cmd.Name(),
			Description: cmd.Description(),
		}
		if cmd.Scope() == command.ScopeUser {
			userScopeBotCommands = append(userScopeBotCommands, bc)
		} else {
			chatScopeBotCommands = append(chatScopeBotCommands, bc)
		}
	}

	if _, err := a.Bot.API().BotsSetBotCommands(context.Background(), &tg.BotsSetBotCommandsRequest{
		Scope:    &tg.BotCommandScopeUsers{},
		Commands: userScopeBotCommands,
	}); err != nil {
		logger.L.Error("bot command users", "error", err)
	}

	if _, err := a.Bot.API().BotsSetBotCommands(context.Background(), &tg.BotsSetBotCommandsRequest{
		Scope:    &tg.BotCommandScopeChats{},
		Commands: chatScopeBotCommands,
	}); err != nil {
		logger.L.Error("bot command chats", "error", err)
	}
}
