package bot

import (
	"activity-bot/internal/command"
	"activity-bot/internal/logger"
	"activity-bot/internal/middleware"
	"context"
	"time"

	"github.com/gotd/td/tg"
)

func (a *App) RegisterHandlers() {
	f := a.createCommandFactory()

	a.registerHelpHandlers(f)
	a.registerChatHandlers(f)
	a.registerAdminHandlers(f)
	a.registerStatsHandlers(f)
	a.registerCallHandlers(f)
	a.registerMemberHandlers(f)
	a.registerUserHandlers(f)
	a.registerRestHandlers(f)
	a.registerMessageHandlers(f)
	a.registerChannelHandlers(f)

	a.setBotCommands(f)
}

func (a *App) createCommandFactory() *command.Factory {
	return command.NewCommandFactory(
		a.UserService,
		a.MemberService,
		a.ChatService,
		a.SessionService,
		a.Config.BotOwnerID,
		"фм", "!", "/", ".",
	)
}

func (a *App) getRateLimiterMiddleware(limit int, seconds int) command.Middleware {
	return middleware.NewRateLimiter(a.Rdb, limit, time.Duration(seconds)*time.Second)
}

func (a *App) setBotCommands(f *command.Factory) {
	var userScopeBotCommands []tg.BotCommand
	var chatScopeBotCommands []tg.BotCommand

	for _, cmd := range f.ConfigurableCommands() {
		if !cmd.Important() {
			continue
		}
		bc := tg.BotCommand{
			Command:     cmd.Name(),
			Description: cmd.Description(),
		}
		if cmd.Scope() == command.ScopeUser {
			userScopeBotCommands = append(userScopeBotCommands, bc)
			chatScopeBotCommands = append(chatScopeBotCommands, bc)
		} else {
			chatScopeBotCommands = append(chatScopeBotCommands, bc)
		}
	}

	if _, err := a.Bot.API().BotsSetBotCommands(context.Background(), &tg.BotsSetBotCommandsRequest{
		Scope:    &tg.BotCommandScopeUsers{},
		Commands: userScopeBotCommands,
	}); err != nil {
		logger.L.Error("bot command users", "error", err)
	}

	if _, err := a.Bot.API().BotsSetBotCommands(context.Background(), &tg.BotsSetBotCommandsRequest{
		Scope:    &tg.BotCommandScopeChats{},
		Commands: chatScopeBotCommands,
	}); err != nil {
		logger.L.Error("bot command chats", "error", err)
	}
}
