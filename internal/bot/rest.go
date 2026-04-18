package bot

import (
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	restH "activity-bot/internal/rest/handler"

	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
)

func (a *App) registerRestHandlers(f *command.Factory) {
	dateParser := helpers.NewDateParser()
	restHandler := restH.New(a.RestService, a.UserService, a.MemberService, a.ChatService, a.AdminService, dateParser, a.SessionService, a.AsyncClient)

	a.dp.AddHandler(f.New("rests_history", restHandler.ShowRest).
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
		SetImportant(true).
		SetArgRules(command.AnyUserRule(), command.OneDateRule()),
	)
	a.dp.AddHandler(f.New("rest", restHandler.ShowRest).SetDescription("Информация о ресте").SetCategory(command.CategoryProfile).
		SetAliases("рест", "rest").
		SetImportant(true).
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
}
