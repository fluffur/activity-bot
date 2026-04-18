package bot

import (
	callH "activity-bot/internal/call/handler"
	"activity-bot/internal/command"
	"activity-bot/internal/conversation"
	"activity-bot/internal/middleware"
	"activity-bot/internal/model"
	"time"

	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
)

func (a *App) registerCallHandlers(f *command.Factory) {
	storage := conversation.NewRedisStorage(a.Rdb, "convo")
	callHandler := callH.New(a.CallService, a.MemberService, a.ChatService, a.AdminService, a.SessionService, storage)

	a.dp.AddHandler(f.New("enable_call_on_join", callHandler.EnableCallOnJoin).
		SetAliases("call_enable", "включить call", "включить колл", "включить калл").
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetDescription("Включить призыв при входе").
		SetCategory(command.CategoryCall),
	)
	a.dp.AddHandler(f.New("set_call_message", callHandler.SetWelcomeCallMessage).
		SetAliases("калл сообщение").
		AddPrefixes("+").
		SetDescription("Настройка сообщения призыва").
		SetCategory(command.CategoryCall).
		SetRequiredStatus(model.StatusSeniorAdmin).
		SetArgRules(command.TextRule()),
	)
	a.dp.AddHandler(f.New("show_call_message", callHandler.ShowWelcomeCallMessage).
		SetAliases("калл сообщение").
		SetProviders(a.UserService, a.MemberService, a.ChatService, a.SessionService).
		SetDescription("Показать сообщение призыва").
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
		SetDescription("Удалить сообщение призыва").
		SetCategory(command.CategoryCall),
	)
	a.dp.AddHandler(f.New("call_type", callHandler.ShowCallTypes).
		SetAliases("калл тип").
		SetImportant(true).
		SetDescription("Показать типы призыва").
		SetCategory(command.CategoryCall),
	)
	a.dp.AddHandler(
		f.New("call_type_callback", callHandler.ShowCallTypes).
			WrapCallback(filters.CallbackQuery.Equal("call_type")),
	)
	a.dp.AddHandler(
		f.New("call_type_select", callHandler.CallbackCallType).
			SetRequiredStatus(model.StatusCoOwner).
			WrapCallback(filters.CallbackQuery.Prefix("call_type:")),
	)

	a.dp.AddHandler(f.New("call_no_norm_warn", callHandler.CallNoNormWarn).
		SetAliases("калл без нормы варн").
		SetRequiredStatus(model.StatusModerator).
		SetDescription("Призыв без нормы с предупреждением").
		SetCategory(command.CategoryCall),
	)
	a.dp.AddHandler(f.New("call_no_norm", callHandler.CallNoNorm).
		SetAliases("калл без нормы", "созвать без нормы").
		SetImportant(true).
		SetRequiredStatus(model.StatusModerator).
		SetDescription("Призыв тех, без нормы").
		SetCategory(command.CategoryCall),
	)

	a.dp.AddHandler(f.New("call_no_norm_ban", callHandler.CallNoNormBan).
		SetAliases("калл без нормы бан").
		SetRequiredStatus(model.StatusModerator).
		SetDescription("Призыв без нормы с баном").
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
		SetImportant(true).
		SetAliases("калл инактив", "калл неактив", "созвать неактивных").
		SetRequiredStatus(model.StatusModerator).
		SetArgRules(command.OptionalVariadicText()).
		SetDescription("Призыв неактивных").
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
		SetImportant(true).
		SetRequiredStatus(model.StatusModerator).
		SetArgRules(command.OptionalVariadicText()).
		SetDescription("Общий сбор чата").
		WithMiddlewares(middleware.NewRateLimiter(a.Rdb, 1, 10*time.Second)).
		SetCategory(command.CategoryCall),
	)

	a.dp.AddHandler(
		f.New("unreg", callHandler.ExcludeMemberFromCall).
			SetAliases("анрег", "-калл").
			SetDescription("Выйти из призыва").
			SetImportant(true).
			SetArgRules(command.AnyUserRule()).
			SetCategory(command.CategoryCall),
	)

	a.dp.AddHandler(
		f.New("reg", callHandler.IncludeMemberInCall).
			SetAliases("рег", "+калл").
			SetDescription("Вернуться в призыв").
			SetArgRules(command.AnyUserRule()).
			SetImportant(true).
			SetCategory(command.CategoryCall),
	)

	a.dp.AddHandler(
		f.New("unregs", callHandler.ListExcludedMembersFromCall).
			SetAliases("анреги").
			SetDescription("Список вышедших из призыва участников").
			SetArgRules(command.AnyUserRule()).
			SetImportant(true).
			SetCategory(command.CategoryCall),
	)
}
