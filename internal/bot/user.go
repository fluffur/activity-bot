package bot

import (
	"activity-bot/internal/command"
	userH "activity-bot/internal/user/handler"
)

func (a *App) registerUserHandlers(f *command.Factory) {
	userHandler := userH.New(a.UserService)

	a.dp.AddHandler(f.New("set_gender", userHandler.SetGender).SetDescription("Установить пол").SetCategory(command.CategoryProfile).
		SetAliases("мой пол", "установить пол").SetScope(command.ScopeUser).
		SetImportant(true).
		SetArgRules(command.TextRule()),
	)
	a.dp.AddHandler(f.New("gender", userHandler.ShowGender).SetDescription("Посмотреть пол").SetCategory(command.CategoryProfile).
		SetAliases("мой пол").SetScope(command.ScopeUser).
		SetImportant(true).
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
}
