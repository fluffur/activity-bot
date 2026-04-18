package bot

import (
	adminH "activity-bot/internal/admin/handler"
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"

	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
)

func (a *App) registerAdminHandlers(f *command.Factory) {
	dateParser := helpers.NewDateParser()
	adminHandler := adminH.New(a.AdminService, a.MemberService, a.ChatService, dateParser, a.AsyncClient, f)

	a.dp.AddHandler(f.New("admins", adminHandler.ListAdmins).SetDescription("Список администрации").SetCategory(command.CategoryGeneral).
		SetImportant(true).
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
		SetArgRules(command.MentionedUserRule(), command.OptionalVariadicText()).
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
			command.OptionalVariadicText(),
		).
		SetDescription("Мут участника").
		SetCategory(command.CategoryModeration),
	)
	a.dp.AddHandler(f.New("warns", adminHandler.ShowWarns).
		SetDescription("Список предупреждений").
		SetCategory(command.CategoryProfile).
		SetArgRules(command.AnyUserRule()).
		SetImportant(true).
		SetAliases("варны", "преды"),
	)
	a.dp.AddHandler(f.New("warnlist", adminHandler.WarnList).
		SetDescription("Предупреждения в чате").
		SetCategory(command.CategoryModeration).
		SetAliases("варнлист", "все преды").
		SetImportant(true))
	a.dp.AddHandler(f.New("warn", adminHandler.Warn).
		SetAliases("варн", "пред").
		SetArgRules(command.MentionedUserRule(), command.OptionalVariadicText()).
		SetRequiredStatus(model.StatusAdmin).
		SetDescription("Предупреждение").
		SetCategory(command.CategoryModeration),
	)
	a.dp.AddHandler(f.New("clear_warns", adminHandler.ClearWarns).
		SetAliases("очистить преды", "очистить варны").
		SetRequiredStatus(model.StatusAdmin).
		SetArgRules(command.MentionedUserRule()).
		SetDescription("Очистить предупреждения").
		SetCategory(command.CategoryModeration),
	)
	a.dp.AddHandler(f.New("max_warns", adminHandler.ShowMaxWarns).
		SetDescription("Максимальное количество предупреждений").
		SetAliases("макс преды", "макс варны").
		SetCategory(command.CategoryModeration))
	a.dp.AddHandler(f.New("set_max_warns", adminHandler.SetMaxWarns).
		SetDescription("Установить лимит предупреждений").
		SetCategory(command.CategoryModeration).
		SetAliases("max_warns", "макс преды", "макс варны").
		AddPrefixes("+").
		SetArgRules(command.NumberRule()).
		SetRequiredStatus(model.StatusCoOwner),
	)
	a.dp.AddHandler(f.New("rights", adminHandler.ToggleRights).
		SetDescription("Управление правами").
		SetCategory(command.CategoryAdmin).
		SetAliases("права", "rights").
		SetDevCommand(true).
		SetArgRules(command.AnyUserRule(), command.NumberRule()),
	)
	a.dp.AddHandler(f.New("update_chats", adminHandler.UpdateChats).
		SetDescription("Обновить кэш чатов").
		SetCategory(command.CategoryAdmin).
		SetDevCommand(true))

	a.dp.AddHandler(f.New("demote", adminHandler.DemoteTgAdmin).
		SetDescription("Разжаловать администратора в Telegram").
		SetCategory(command.CategoryAdmin).
		SetAliases("разжаловать").
		SetArgRules(command.MentionedUserRule()).
		SetRequiredStatus(model.StatusSeniorAdmin))

	a.dp.AddHandler(f.New("manage_rights", adminHandler.ManageRights).
		SetAliases("дк").
		SetRequiredStatus(model.StatusCoOwner).
		SetDescription("Управление доступом команд").
		SetCategory(command.CategoryAdmin),
	)
	a.dp.AddHandler(
		f.New("manage_rights_callback", adminHandler.CallbackManageRights).
			SetRequiredStatus(model.StatusCoOwner).
			WrapCallback(filters.CallbackQuery.Prefix("rights_")),
	)
}
