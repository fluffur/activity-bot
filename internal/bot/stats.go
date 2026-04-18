package bot

import (
	"activity-bot/internal/command"
	statsH "activity-bot/internal/stats/handler"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
)

func (a *App) registerStatsHandlers(f *command.Factory) {
	statsHandler := statsH.New(a.StatsService, a.RestService, a.MemberService, a.UserService, a.ChatService, a.SessionService)

	a.dp.AddHandler(f.New("stats", statsHandler.ShowStats).
		SetAliases("отчёт", "отчет", "стата").
		SetImportant(true).
		SetArgRules(command.OptionalDateRangeRule()).
		SetDescription("Отчёт в чате").
		SetCategory(command.CategoryStats),
	)
	a.dp.AddHandler(f.New("inactive", statsHandler.ListInactive).
		SetAliases("неактив", "инактив").
		SetDescription("Список неактивных участников сроком более суток").
		SetCategory(command.CategoryStats),
	)
	a.dp.AddHandler(f.New("stats_graph", statsHandler.ShowChatActivityGraph).SetDescription("График активности чата").SetCategory(command.CategoryStats).
		SetAliases("график").
		SetArgRules(command.OptionalDateRangeRule()),
	)
	a.dp.AddHandler(f.New("rests", statsHandler.ShowRestList).
		SetDescription("Список участников в ресте").
		SetImportant(true).
		SetAliases("ресты").
		SetCategory(command.CategoryStats),
	)

	a.dp.AddHandler(f.New("who_am_i", statsHandler.WhoAmI).
		SetDescription("Мой профиль").
		SetImportant(true).
		SetCategory(command.CategoryProfile).
		SetAliases("ктоя", "кто я", "профиль"),
	)
	a.dp.AddHandler(f.New("who_are_u", statsHandler.WhoAreYou).
		SetImportant(true).
		SetDescription("Профиль участника").
		SetCategory(command.CategoryProfile).
		SetArgRules(command.MentionedUserRule()).
		SetAliases("ктоты", "кто ты", "профиль"),
	)
	a.dp.AddHandler(f.New("callback_all_activity", statsHandler.CallbackAllActivity).
		WrapCallback(filters.CallbackQuery.Prefix("profile_activity:")),
	)
	a.dp.AddHandler(f.New("callback_profile_graph", statsHandler.CallbackProfileGraph).
		WrapCallback(filters.CallbackQuery.Prefix("profile_graph:")),
	)
	a.dp.AddHandler(f.New("failed_norm", statsHandler.ShowFailedNorm).
		SetAliases("без нормы").
		SetDescription("Список участников без нормы").
		SetImportant(true).
		SetCategory(command.CategoryStats))
}
