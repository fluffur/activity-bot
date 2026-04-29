package bot

import (
	chatH "activity-bot/internal/chat/handler"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"

	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
)

func (a *App) registerChatHandlers(f *command.Factory) {
	dateParser := helpers.NewDateParser()
	chatHandler := chatH.New(a.ChatService, a.AdminService, a.MemberService, a.SessionService, a.RPService, dateParser)

	a.dp.AddHandler(f.New("set_newbie_treshold", chatHandler.SetNewbieThreshold).
		SetAliases("новички срок", "новички после").
		AddPrefixes("+").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.NumberRule()).
		SetDescription("Настройка срока новичка").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("norm", chatHandler.ShowNorm).
		SetDescription("Посмотреть норму сообщений").
		SetImportant(true).
		SetCategory(command.CategorySettings).
		SetAliases("норма какая", "а норма какая", "норма", "норма?", "quota", "какая норма", "а какая норма"),
	)
	a.dp.AddHandler(f.New("set_norm", chatHandler.SetNorm).
		SetDescription("Установка нормы").
		SetImportant(true).
		SetCategory(command.CategorySettings).
		SetAliases("норма", "quota").
		AddPrefixes("+").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.NumberRule(), command.OptionalVariadicText()),
	)
	a.dp.AddHandler(f.New("remove_norm", chatHandler.RemoveNorm).
		SetDescription("Удалить норму").
		SetCategory(command.CategorySettings).
		SetAliases("норма", "quota").
		AddPrefixes("-").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.OptionalVariadicText()),
	)

	a.dp.AddHandler(f.New("chat_title_change", chatHandler.OnNewChatTitle).WrapEvent(chatTitleChangedFilter))

	a.dp.AddHandler(f.New("manage", chatHandler.Manage).
		SetDescription("Управление чатом").
		SetCategory(command.CategorySettings).
		SetArgRules(command.OptionalVariadicText()).
		SetAliases("управление").
		SetImportant(true).
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
		SetImportant(true).
		SetDescription("Включение тегов").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("disable_tags", chatHandler.DisableTags).
		SetAliases("-tags", "-теги", "-тэги").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Отключение тегов").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("tags", chatHandler.ShowTags).
		SetDescription("Статус тегов").
		SetCategory(command.CategoryGeneral).
		SetAliases("tags", "теги", "тэги"))

	a.dp.AddHandler(f.New("my_norms", chatHandler.UserChats).
		SetDescription("Список норм во всех чатах").SetCategory(command.CategoryGeneral).
		SetAliases("чаты", "нормы", "чаты без нормы"))

	a.dp.AddHandler(f.New("set_prompt", chatHandler.SetPrompt).
		SetAliases("промпт").
		AddPrefixes("+").
		SetArgRules(command.TextRule().SetVariadic(true)).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Настройка промпта для ИИ").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("week_start", chatHandler.ShowWeekStart).SetDescription("Начало недели для нормы").SetCategory(command.CategoryGeneral).
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
		SetCategory(command.CategoryGeneral).
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
	a.dp.AddHandler(f.New("show_newbie_treshold", chatHandler.ShowNewbieThreshold).SetDescription("Порог новичка").SetCategory(command.CategoryGeneral).
		SetAliases("новички срок").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Показать срок новичка").
		SetCategory(command.CategoryGeneral),
	)

	a.dp.AddHandler(f.New("prompt", chatHandler.ShowPrompt).
		SetDescription("Системный ИИ промпт").
		SetCategory(command.CategoryGeneral).
		SetAliases("промпт").
		SetDescription("Показать текущий AI-промпт").
		SetCategory(command.CategoryGeneral),
	)

	a.dp.AddHandler(f.New("prefix_only", chatHandler.ShowPrefixlessStatus).
		SetAliases("префиксы").
		SetImportant(true).
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
		SetAliases("-префиксы", "без префиксов").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Отключить обязательный префикс").
		SetCategory(command.CategorySettings),
	)
	a.dp.AddHandler(f.New("rp_preview", chatHandler.PreviewRPTemplate).
		SetAliases("рп превью").
		SetArgRules(command.OptionalVariadicText()).
		SetDescription("Превью РП-шаблона").
		SetCategory(command.CategoryGeneral),
	)

	a.dp.AddHandler(f.New("set_rp_command", chatHandler.SetRPCommand).
		SetAliases("+рп").
		SetArgRules(command.OptionalVariadicText()).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Добавить РП-команду").
		SetCategory(command.CategorySettings),
	)

	a.dp.AddHandler(f.New("remove_rp_command", chatHandler.RemoveRPCommand).
		SetAliases("-рп").
		SetArgRules(command.OptionalVariadicText()).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Удалить РП-команду").
		SetCategory(command.CategorySettings),
	)

	a.dp.AddHandler(f.New("rp_commands", chatHandler.ListRPCommands).
		SetAliases("рп").
		SetDescription("Список РП-команд").
		SetCategory(command.CategoryGeneral),
	)

}
