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
	"activity-bot/internal/session"
	"activity-bot/internal/stats"
	statsH "activity-bot/internal/stats/handler"
	userH "activity-bot/internal/user/handler"
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
	sessionRepository := postgres.NewSessionRepository(queries)

	statsService := stats.NewService(statsRepository)
	restService := rest.NewService(restRepository)
	messageService := msg.NewService(messageRepository)

	callService := call.NewService(chatRepository, a.MemberService)
	sessionService := session.NewService(sessionRepository)

	dateParser := helpers.NewDateParser()
	cf := cmd.NewFactory(a.UserService, a.ChatService, a.MemberService, sessionService, a.Config.UniquePrefix, "/", "!", ".")

	helpHandler := helpH.New(a.Config.BotOwnerID)
	statsHandler := statsH.New(statsService, restService, a.MemberService, a.UserService, a.ChatService, sessionService)
	chatHandler := chatH.New(a.ChatService, a.MemberService, a.AdminService, sessionService, dateParser, cf)
	restHandler := restH.New(restService, a.UserService, a.MemberService, a.AdminService, dateParser, sessionService, a.AsyncClient)

	adminHandler := adminH.New(a.AdminService, a.UserService, a.MemberService, a.ChatService, dateParser, a.AsyncClient)

	messageHandler := messageH.New(messageService, a.MemberService, a.ChatService, a.Deepseek)
	memberHandler := memberH.New(a.MemberService, a.ChatService, a.UserService, callService, a.AdminService)
	callHandler := callH.New(callService, a.ChatService, a.AdminService, sessionService)
	userHandler := userH.New(a.UserService)

	dynamicLevelGuard := guard.NewLevelGuard(a.AdminService, a.MemberService, sessionService, cf, 0)
	cf.AddDefaultGuards(dynamicLevelGuard)

	adminGuard := guard.NewLevelGuard(a.AdminService, a.MemberService, sessionService, nil, 3)
	creatorGuard := guard.NewLevelGuard(a.AdminService, a.MemberService, sessionService, nil, 5)
	ownerGuard := guard.NewDevCreatorGuard(a.AdminService, sessionService)
	developerGuard := guard.NewDeveloperGuard(a.AdminService, sessionService)
	groupGuard := guard.OnlyGroups(sessionService)
	rateLimiterGuard := guard.NewRateLimiter(a.Rdb, 2, 10*time.Second, sessionService)

	a.Dispatcher.AddHandler(cf.New(helpHandler.Start, "start").
		SetID("start").
		SetDescription("Начать работу с ботом").
		SetCategory("Полезное").
		SetLevel(0))
	a.Dispatcher.AddHandler(cf.New(helpHandler.Help, "help").
		SetID("help").
		SetDescription("Показать список команд").
		SetCategory("Полезное").
		SetLevel(0))
	a.Dispatcher.AddHandler(cf.New(callHandler.ShowWelcomeCallMessage, "call_message", "call сообщение", "колл сообщение", "калл сообщение").
		SetDescription("Показать приветственное сообщение для каллов").
		SetCategory("Активность").
		SetLevel(0).
		AddTriggers("+").
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.SetWelcomeCallMessage, "call_message", "call сообщение", "колл сообщение", "калл сообщение").
		SetID("call_message_set").
		SetDescription("Установить приветственное сообщение для каллов").
		SetCategory("Активность").
		SetLevel(3).
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)

	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowNewbieThreshold, "newbie_threshold_show", "новички срок", "новички после").
		SetID("newbie_threshold_show").
		SetDescription("Показать порог для новичков").
		SetCategory("Активность").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetNewbieThreshold, "newbie_threshold_set", "новички срок", "новички после").
		SetID("newbie_threshold_set").
		SetDescription("Установить порог для новичков (в днях)").
		SetDetailedDescription("Можно указать срок, в течение которого участник считается новичком. По истечении этого срока участник переходит в категорию старичков.").
		SetExamples("новички после 7 дней", "новички после").
		SetCategory("Активность").
		SetLevel(3).
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ShowStats, "stats", "отчёт", "отчет", "стата").
		SetID("stats").
		SetDescription("Показать статистику участников").
		SetDetailedDescription("Команда для получения статистики по сообщениям всех участников с разными категориями. По умолчанию отображается активность за текущую календарную неделю.").
		SetExamples("отчет", "отчет 1-10", "отчет 7", "отчет 01.01-04.04").
		SetCategory("Активность").
		SetLevel(0).
		SetArgsCount(1).
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 2, 4*time.Second, sessionService)),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ShowChatActivityGraph, "stats_graph", "график", "граф").
		SetID("stats_graph").
		SetDescription("Показать график активности чата").
		SetCategory("Активность").
		SetLevel(0).
		SetArgsCount(1).
		WithGuards(groupGuard, rateLimiterGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.WhoAreYou, "whoareu", "ктоты", "кто ты", "профиль", "ты кто", "тыкто").
		SetID("whoareu").
		SetDescription("Показать профиль другого участника").
		SetDetailedDescription("Команда для получения информации о пользователе, включая статистику активности. Можно посмотреть профиль участника через упоминание, ссылку или название роли.").
		SetExamples("профиль", "профиль @yoworu", "профиль https://t.me/yoworu", "профиль яблоня").
		SetCategory("Активность").
		SetLevel(0).
		SetArgsCount(1).
		WithGuards(groupGuard).FallbackToSender(),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.WhoAmI, "whoami", "кто я", "профиль", "ктоя", "я кто").
		SetID("whoami").
		SetDescription("Показать свой профиль").
		SetCategory("Активность").
		SetLevel(0).
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.WhoAmI, "я", "me").
		SetID("me").
		SetDescription("Показать свой профиль (коротко)").
		SetCategory("Активность").
		SetLevel(0).
		ForcePrefix().
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("whoareyou:"), cf.WrapCallback(statsHandler.CallbackWhoAreYou)))
	a.Dispatcher.AddHandler(cf.New(statsHandler.Inactive, "inactive", "неактив", "инактив").
		SetID("inactive").
		SetDescription("Показать список неактивных участников").
		SetCategory("Активность").
		SetLevel(3).
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ShowRestList, "rests", "ресты").
		SetID("rests").
		SetDescription("Показать список участников в ресте").
		SetCategory("Активность").
		SetLevel(0).
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ShowFailedNorm, "nonorm", "без нормы").
		SetID("nonorm").
		SetDescription("Показать список участников, не выполнивших норму").
		SetCategory("Активность").
		SetLevel(0).
		SetArgsCount(1).
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 2, 4*time.Second, sessionService)),
	)
	a.Dispatcher.AddHandler(cf.New(statsHandler.ShowNewbies, "newbies", "новички").
		SetID("newbies_list").
		SetDescription("Показать список новичков чата").
		SetCategory("Активность").
		SetLevel(0).
		SetArgsCount(1).
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 2, 4*time.Second, sessionService)),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowNorm, "norm", "норма какая", "а норма какая", "норма", "норма?", "quota", "какая норма", "а какая норма").
		SetID("norm_show").
		SetDescription("Показать текущую норму чата").
		SetCategory("Активность").
		SetLevel(0).
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetNorm, "norm", "норма", "quota").
		SetID("norm_set").
		SetDescription("Установить норму сообщений").
		SetDetailedDescription("Норма - минимальное число сообщений в неделю, которое должен набрать участник чата. Нужна для команд 'отчет' и 'профиль'.").
		SetExamples("+норма 150", "+норма 100 бан", "норма").
		SetCategory("Активность").
		SetLevel(3).
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(2),
	)
	a.Dispatcher.AddHandler(cf.New(restHandler.Show, "рест", "rest", "мой рест").
		SetID("rest_show").
		SetDescription("Показать информацию о вашем ресте").
		SetCategory("Активность").
		SetLevel(0).
		FallbackToSender().
		WithGuards(groupGuard).
		AddTriggers("+"),
	)
	a.Dispatcher.AddHandler(cf.New(restHandler.Set, "рест", "rest", "установить рест").
		SetID("rest_set").
		SetDescription("Установить рест (причина и срок)").
		SetDetailedDescription("Рест - это время, в течение которого участник освобожден от выполнения нормы. Можно добавлять участников в рест, если вы администратор, и отправлять заявку на рест, если вы участник.").
		SetExamples("рест неделя", "рест 8 апреля", "рест 12 августа 2036", "рест 01.01.2077").
		SetCategory("Активность").
		SetLevel(0).
		FallbackToSender().
		AddTriggers("+").
		WithGuards(groupGuard).
		SetArgsCount(2),
	)
	a.Dispatcher.AddHandler(cf.New(restHandler.End, "-рест", "-rest", "снять рест").
		SetID("rest_end").
		SetDescription("Завершить рест").
		SetCategory("Активность").
		SetLevel(0).
		FallbackToSender().
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ListAdmins, "admins", "админы", "админчики", "администраторы", "адмы", "модеры", "mods").
		SetID("admins_list").
		SetDescription("Показать список администраторов чата").
		SetDetailedDescription("Просмотр списка всех пользователей, имеющих статус администратора в данном чате.").
		SetExamples("админы").
		SetCategory("Модерация").
		SetLevel(0).
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 1, 10*time.Second, sessionService)),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.IsAdmin, "админ", "admin", "is_admin", "адм", "модер", "mod", "is_mod").
		SetID("admin_check").
		SetDescription("Проверить наличие прав администратора").
		SetCategory("Модерация").
		SetLevel(0).
		WithGuards(groupGuard).
		FallbackToSender(),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.AddAdmin, "+админ", "+admin", "+адм", "+модер", "+mod").
		SetID("admin_add").
		SetDescription("Добавить администратора").
		SetCategory("Администрирование").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.RemoveAdmin, "-администратор", "-админ", "-admin", "-адм", "-модер", "-mod").
		SetID("admin_remove").
		SetDescription("Удалить администратора").
		SetCategory("Администрирование").
		SetLevel(5).
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Unban, "unban", "-бан", "разбан", "разбанить").
		SetID("unban").
		SetDescription("Разбанить пользователя").
		SetCategory("Модерация").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Unmute, "unmute", "размут", "размутить", "-мут").
		SetID("unmute").
		SetDescription("Размутить пользователя").
		SetCategory("Модерация").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Unwarn, "unwarn", "анварн", "снятьпред", "-варн", "-пред").
		SetID("unwarn").
		SetDescription("Снять предупреждение").
		SetCategory("Модерация").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard),
	)

	a.Dispatcher.AddHandler(cf.New(adminHandler.Kick, "kick", "кик", "выгнать").
		SetID("kick").
		SetDescription("Кикнуть пользователя").
		SetCategory("Модерация").
		SetLevel(3).
		AddTriggers("+").
		SetArgsCount(1).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Ban, "ban", "бан").
		SetID("ban").
		SetDescription("Забанить пользователя").
		SetDetailedDescription("Блокировка доступа в чат. Можно указать срок блокировки.").
		SetExamples("бан @user 7 дней").
		SetCategory("Модерация").
		SetLevel(3).
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(2),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Mute, "mute", "мут", "замутить").
		SetID("mute").
		SetDescription("Замутить пользователя").
		SetDetailedDescription("Ограничение на отправку сообщений в чат. Можно указать срок мута.").
		SetExamples("мут @user час").
		SetCategory("Модерация").
		SetLevel(3).
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(2),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ShowWarns, "warns", "варны", "преды").
		SetID("warns_show").
		SetDescription("Показать предупреждения пользователя").
		SetCategory("Модерация").
		SetLevel(0).
		FallbackToSender().
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Warnlist, "warnlist", "варнлист", "предывсе").
		SetID("warnlist").
		SetDescription("Показать список всех предупреждений в чате").
		SetCategory("Модерация").
		SetLevel(3).
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.Warn, "warn", "варн", "пред", "предупреждение").
		SetID("warn").
		SetDescription("Выдать предупреждение").
		SetDetailedDescription("Выдача предупреждения пользователю. При достижении лимита (макс варны) последует автоматический бан.").
		SetExamples("варн @user час", "пред").
		SetCategory("Модерация").
		SetLevel(3).
		AddTriggers("+").SetArgsCount(2).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ClearWarns, "clear_warns", "очистить преды", "очистить варны").
		SetID("warns_clear").
		SetDescription("Очистить все предупреждения пользователя").
		SetCategory("Модерация").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ShowMaxWarns, "макс преды", "макс варны", "max_warns").
		SetID("max_warns_show").
		SetDescription("Показать максимальное количество предупреждений").
		SetCategory("Настройки").
		SetLevel(0).
		WithGuards(groupGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.SetMaxWarns, "max_warns", "макс преды", "макс варны").
		SetID("max_warns_set").
		SetDescription("Установить максимальное количество предупреждений").
		SetDetailedDescription("Лимит предупреждений, после которого следует автоматический бан пользователя.").
		SetExamples("макс варны 5").
		SetCategory("Настройки").
		SetLevel(5).
		AddTriggers("+").
		SetArgsCount(1).
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ToggleRights, "права", "rights").
		SetID("rights_toggle").
		SetDescription("Переключить права администратора").
		WithGuards(developerGuard).SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.AddDeveloper, "дев", "adddev").
		SetID("dev_add").
		SetDescription("Добавить разработчика").
		WithGuards(ownerGuard).SetArgsCount(1).
		AddTriggers("+"),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.RemoveDeveloper, "-дев", "remdev").
		SetID("dev_remove").
		SetDescription("Удалить разработчика").
		WithGuards(ownerGuard).SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.ListDevelopers, "девс", "devs").
		SetID("devs_list").
		SetDescription("Показать список разработчиков").
		WithGuards(developerGuard),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.UpdateChats, "update_chats").
		SetID("update_chats").
		SetDescription("Обновить данные о чатах").
		WithGuards(developerGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.UpdateMembersList, "обновить чат", "update chat", "update").
		SetID("update_chat_members").
		SetDescription("Обновить список участников чата").
		SetCategory("Полезное").
		SetLevel(3).
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 1, 10*time.Second, sessionService)),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.ListRoles, "роли", "roles", "titles").
		SetID("roles_list").
		SetDescription("Показать список ролей").
		SetCategory("Роли").
		SetLevel(0).
		WithGuards(groupGuard, rateLimiterGuard),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.ShowRole, "роль", "role", "title",
		"какая роль", "роль у", "роль кого").
		SetID("role_show").
		SetDescription("Показать роль участника").
		SetCategory("Роли").
		SetLevel(0).
		WithGuards(groupGuard).
		FallbackToSender(),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.SetRole, "роль", "role", "title").
		SetID("role_set").
		SetDescription("Установить роль участнику").
		SetDetailedDescription("Установка или изменение роли (подписи) участника. Бот поддерживает как выдачу прав администратора с подписью, так и использование тегов Телеграм.").
		SetExamples("!роль @участник название роли").
		SetCategory("Роли").
		SetLevel(3).
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(memberHandler.RestoreRoles, "восстановить роли", "restore_roles").
		SetID("roles_restore").
		SetDescription("Восстановить роли участникам").
		SetCategory("Роли").
		SetLevel(5).
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.ShowCallTypes, "call_type", "калл тип", "калл стиль").
		SetID("call_type_show").
		SetDescription("Показать доступные типы звонков").
		SetCategory("Звонки").
		SetLevel(3).
		AddTriggers("+", "!").
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.SetMentionsPerMessage, "call_limit", "калл лимит", "калл лим").
		SetID("call_limit_set").
		SetDescription("Установить лимит упоминаний в одном сообщении").
		SetCategory("Звонки").
		SetLevel(3).
		AddTriggers("+", "!").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.Call, "call", "калл", "колл", "all", "каллалл").
		SetID("call_all").
		SetDescription("Позвать всех участников").
		SetDetailedDescription("Инструмент созыва участников чата. Можно добавить сообщение, которое будет отправлено вместе с упоминаниями.").
		SetExamples("калл сообщение", "калл").
		SetCategory("Звонки").
		SetLevel(3).
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard, rateLimiterGuard).
		SetArgsCount(1),
	)

	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("call_type:"), cf.WrapCallback(callHandler.CallbackCallType)))
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowPrompt, "промпт").
		SetID("prompt_show").
		SetDescription("Показать системный промпт AI").
		SetCategory("Настройки").
		SetLevel(3).
		WithGuards(groupGuard))
	a.Dispatcher.AddHandler(handlers.NewMessage(cmd.NewChatTitle, cf.WrapEvent(chatHandler.OnNewChatTitle)))
	a.Dispatcher.AddHandler(cf.New(chatHandler.Manage, "manage", "управление").
		SetID("manage").
		SetDescription("Управление чатами в ЛС").
		SetDetailedDescription("Написав боту в ЛС команду 'управление', бот выдаст список чатов, которыми управляет пользователь. Выбрав нужный чат, можно вызывать команды, относящиеся к этому чату, в ЛС у бота.").
		SetExamples("управление").
		SetCategory("Общее").
		SetLevel(0))
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("manage:"), cf.WrapCallback(chatHandler.CallbackManage)))
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("manage_page:"), cf.WrapCallback(chatHandler.CallbackManagePage)))
	a.Dispatcher.AddHandler(cf.New(chatHandler.EnableTags, "+tags", "+теги", "+тэги").
		SetID("tags_enable").
		SetDescription("Включить поддержку тегов").
		SetCategory("Настройки").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard))
	a.Dispatcher.AddHandler(cf.New(chatHandler.DisableTags, "-tags", "-теги", "-тэги").
		SetID("tags_disable").
		SetDescription("Выключить поддержку тегов").
		SetCategory("Настройки").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard))
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowTags, "tags", "теги", "тэги").
		SetID("tags_show").
		SetDescription("Показать статус поддержки тегов").
		SetDetailedDescription("С 1 марта 2026 года Телеграм добавил возможность устанавливать подписи пользователям без необходимости выдавать статус администратора. Такие подписи называются тегами. Бот по умолчанию при выдаче роли участнику использует теги.").
		SetExamples("теги").
		SetCategory("Настройки").
		SetLevel(0))
	a.Dispatcher.AddHandler(cf.New(chatHandler.UserChats, "chats", "чаты", "нормы", "чаты без нормы").
		SetID("user_chats").
		SetDescription("Показать список ваших чатов и норм").
		SetCategory("Общее").
		SetLevel(0))
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetPrompt, "промпт").
		SetID("prompt_set").
		SetDescription("Установить системный промпт AI").
		SetCategory("Настройки").
		SetLevel(3).
		AddTriggers("+").
		SetArgsCount(1).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetCommandLevel, "дк").
		SetID("set_cmd_level").
		SetDescription("Установить требуемый уровень для команды").
		SetCategory("Настройки").
		SetLevel(5).
		SetArgsCount(2).
		WithGuards(groupGuard, creatorGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowWeekStart, "week_start", "начало недели", "чистка", "время чистки", "конец чистки").
		SetID("week_start_show").
		SetDescription("Показать время начала чистки").
		SetCategory("Настройки").
		SetLevel(0))
	a.Dispatcher.AddHandler(cf.New(chatHandler.ManageWeekStart, "week_start", "начало недели", "чистка", "время чистки", "конец чистки").
		SetID("week_start_set").
		SetDescription("Установить время начала чистки").
		SetDetailedDescription("Можно указать конец чистки (начало недели, от которого будет идти следующая неделя). Время должно быть указано по Московскому времени (МСК).").
		SetExamples("!чистка воскресенье 18:00").
		SetCategory("Настройки").
		SetLevel(3).
		AddTriggers("+").
		SetArgsCount(1).WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowPrefix, "custom_prefix", "кастом префикс", "префикс").
		SetID("prefix_show").
		SetDescription("Показать текущий префикс команд").
		SetCategory("Настройки").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.SetPrefix, "custom_prefix", "кастом префикс", "префикс").
		SetID("prefix_set").
		SetDescription("Установить префикс команд").
		SetDetailedDescription("Можно установить дополнительный префикс для бота. По умолчанию бот реагирует на /, ! и .").
		SetExamples("кастом префикс тотя", "кастом префикс").
		SetCategory("Настройки").
		SetLevel(3).
		AddTriggers("+").
		WithGuards(groupGuard, adminGuard).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.ShowPrefixes, "префиксы", "prefixes").
		SetID("prefixless_show").
		SetDescription("Показать статус работы без префиксов").
		SetCategory("Настройки").
		SetLevel(0).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.DisablePrefixes, "с префиксами", "-prefixless").
		SetID("prefixless_disable").
		SetDescription("Отключить работу без префиксов").
		SetCategory("Настройки").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(chatHandler.EnablePrefixes, "без префиксов", "+prefixless").
		SetID("prefixless_enable").
		SetDescription("Включить работу без префиксов").
		SetCategory("Настройки").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.EnableCallOnJoin, "call_enable", "включить call", "включить колл", "включить калл").
		SetID("call_on_join_enable").
		SetDescription("Включить автоматический звонок при вступлении").
		SetCategory("Звонки").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(callHandler.DisableCallOnJoin, "call_disable", "отключить call", "отключить колл", "отключить калл").
		SetID("call_on_join_disable").
		SetDescription("Отключить автоматический звонок при вступлении").
		SetCategory("Звонки").
		SetLevel(3).
		WithGuards(groupGuard, adminGuard),
	)
	a.Dispatcher.AddHandler(cf.New(messageHandler.Bot, "крис").
		SetID("ai_chat").
		SetDescription("Общение с AI").
		SetCategory("Общее").
		SetLevel(0).
		WithGuards(groupGuard, guard.NewRateLimiter(a.Rdb, 5, 10*time.Second, sessionService)).
		SetArgsCount(1),
	)
	a.Dispatcher.AddHandler(cf.New(adminHandler.DemoteTgAdmin, "разжаловать").
		SetID("admin_demote").
		SetDescription("Снять права администратора (Telegram)").
		SetCategory("Администрирование").
		SetLevel(5).
		WithGuards(groupGuard, adminGuard))
	a.Dispatcher.AddHandler(cf.New(adminHandler.FakeLeave, "фейклив").
		SetID("fake_leave").
		SetDescription("Фейковый выход из чата").
		SetCategory("Общее").
		SetLevel(3).
		FallbackToSender().WithGuards(groupGuard, adminGuard))
	a.Dispatcher.AddHandler(cf.New(userHandler.ShowGender, "пол", "gender").
		SetID("gender_show").
		SetDescription("Показать ваш пол").
		SetCategory("Общее").
		SetLevel(0).
		FallbackToSender())
	a.Dispatcher.AddHandler(cf.New(userHandler.SetGender, "пол", "gender").
		SetID("gender_set").
		SetDescription("Установить ваш пол").
		SetCategory("Общее").
		SetLevel(0).
		FallbackToSender().
		SetArgsCount(1),
	)

	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("approve:"), cf.WrapCallback(restHandler.ApproveRestRequest)))
	a.Dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("reject:"), cf.WrapCallback(restHandler.RejectRestRequest)))
	a.Dispatcher.AddHandler(handlers.NewMessage(message.LeftChatMember, cf.WrapEvent(memberHandler.OnLeftMember)))
	a.Dispatcher.AddHandler(handlers.NewMessage(message.NewChatMembers, cf.WrapEvent(memberHandler.OnJoinMember)))
	a.Dispatcher.AddHandler(handlers.NewMyChatMember(chatmember.NewStatus("administrator"), cf.WrapEvent(memberHandler.OnBotPromote)))
	a.Dispatcher.AddHandler(handlers.NewMessage(filter.OnlyGroups, cf.WrapEvent(messageHandler.Message)))
}
