package bot

import (
	channelH "activity-bot/internal/channel/handler"
	"activity-bot/internal/command"
	helpH "activity-bot/internal/help/handler"
	messageH "activity-bot/internal/message/handler"
	"activity-bot/internal/model"

	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
)

func (a *App) registerHelpHandlers(f *command.Factory) {
	helpHandler := helpH.New(a.Config.BotOwnerUsername, a.Config.CommandsLink)

	a.dp.AddHandler(f.New("start", helpHandler.Start).
		SetScope(command.ScopeUser),
	)
	a.dp.AddHandler(f.New("help", helpHandler.Help).
		SetScope(command.ScopeUser).
		SetImportant(true).
		SetCategory(command.CategorySettings).
		SetDescription("Помощь по боту"),
	)
}

func (a *App) registerMessageHandlers(f *command.Factory) {
	messageHandler := messageH.New(a.MessageService, a.MemberService, a.ChatService, a.RPService, a.Deepseek)

	a.dp.AddHandler(f.New("ask_ai", messageHandler.Bot).
		SetDescription("Вопрос к ИИ").
		SetCategory(command.CategoryFun).
		SetAliases("крис").
		SetArgRules(command.OptionalVariadicText()),
	)

	a.dp.AddHandlerToGroup(
		f.New("rp_event", messageHandler.HandleRPCommand).
			SetArgRules(command.MentionedUserRule()).WrapEvent(textMessageFilter), 1,
	)
	a.dp.AddHandlerToGroup(
		f.New("message", messageHandler.Message).WrapEvent(textMessageFilter), 1,
	)
}

func (a *App) registerChannelHandlers(f *command.Factory) {
	channelHandler := channelH.New(a.MemberService, a.ChatService, a.AsyncClient, a.Config.ChannelID)

	a.dp.AddHandler(
		f.New("subscribe", channelHandler.Subscribe).SetDescription("Подписка на канал").SetCategory(command.CategorySettings).
			SetRequiredStatus(model.StatusSeniorAdmin),
	)
	a.dp.AddHandler(
		f.New("unsubscribe", channelHandler.Unsubscribe).SetDescription("Отписка от канала").SetCategory(command.CategorySettings).
			WrapCallback(filters.CallbackQuery.Prefix("unsubscribe")),
	)
}
