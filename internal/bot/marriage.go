package bot

import (
	"activity-bot/internal/command"
	"activity-bot/internal/marriage/handler"

	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
)

func (a *App) registerMarriageHandlers(f *command.Factory) {
	marriageHandler := handler.New(a.MarriageService, a.MemberService)

	a.dp.AddHandler(f.New("marry", marriageHandler.RequestMarriage).
		SetDescription("Сделать предложение брака").
		SetCategory(command.CategoryProfile).
		SetAliases("брак запрос", "пожениться", "брак").
		SetArgRules(command.AnyUserRule()),
	)

	a.dp.AddHandler(f.New("marriage", marriageHandler.ShowMarriage).
		SetDescription("Показать активный брак").
		SetCategory(command.CategoryProfile).
		SetAliases("мой брак").
		SetArgRules(command.AnyUserRule()),
	)

	a.dp.AddHandler(f.New("marriages", marriageHandler.ListMarriages).
		SetDescription("Список активных браков").
		SetCategory(command.CategoryProfile).
		SetAliases("браки"),
	)

	a.dp.AddHandler(f.New("divorce", marriageHandler.Divorce).
		SetDescription("Оформить развод").
		SetCategory(command.CategoryProfile).
		SetAliases("развод").
		SetArgRules(command.AnyUserRule()),
	)

	a.dp.AddHandler(
		f.New("marriage_accept", marriageHandler.AcceptMarriageRequest).
			WrapCallback(filters.CallbackQuery.Prefix("marriage_accept:")),
	)

	a.dp.AddHandler(
		f.New("marriage_reject", marriageHandler.RejectMarriageRequest).
			WrapCallback(filters.CallbackQuery.Prefix("marriage_reject:")),
	)
}
