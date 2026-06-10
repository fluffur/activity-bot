package bot

import (
	"activity-bot/internal/command"
	memberH "activity-bot/internal/member/handler"
	"activity-bot/internal/model"
)

func (a *App) registerMemberHandlers(f *command.Factory) {
	memberHandler := memberH.New(a.MemberService, a.ChatService, a.UserService, a.CallService, a.AdminService)

	a.dp.AddHandler(
		f.New("new_members", memberHandler.OnJoinMember).WrapEvent(joinMemberFilter),
	)

	a.dp.AddHandler(f.New("ship_random", memberHandler.ShipRandom).
		SetDescription("Случайный шип").
		SetImportant(true).
		SetCategory(command.CategoryFun).
		SetAliases("рандом шипперим", "шипперим"))

	a.dp.AddHandler(f.New("update_chat", memberHandler.UpdateMembersList).SetDescription("Обновление списка участников").SetCategory(command.CategoryAdmin).
		SetAliases("обновить чат", "update").
		WithMiddlewares(a.getRateLimiterMiddleware(3, 10)).
		SetRequiredStatus(model.StatusMember).
		SetDescription("Обновление списка участников").
		SetCategory(command.CategoryAdmin),
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
		SetCategory(command.CategoryAdmin),
	)
	a.dp.AddHandler(f.New("role", memberHandler.ShowRole).SetDescription("Посмотреть роль").
		SetArgRules(command.AnyUserRule()).
		SetCategory(command.CategoryProfile).
		SetAliases("роль", "title", "какая роль", "роль у", "роль кого"),
	)
	a.dp.AddHandler(f.New("set_member_emoji", memberHandler.SetEmoji).SetAliases("значок").
		SetArgRules(command.AnyUserRule(), command.TextRule()).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Настройка значка участника").
		SetCategory(command.CategoryProfile),
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

	a.dp.AddHandler(f.New("fake_leave", memberHandler.FakeLeave).
		SetDescription("Сымитировать выход из чата").
		SetCategory(command.CategoryFun).
		SetAliases("фейклив", "фейк лив").
		SetArgRules(command.AnyUserRule()))

	a.dp.AddHandler(
		f.New("left_member", memberHandler.OnLeftMember).WrapEvent(leftMemberFilter),
	)
}
