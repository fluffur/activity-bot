package guard

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type DeveloperGuard struct {
	service        *admin.Service
	sessionService cmd.SessionService
}

func NewDeveloperGuard(service *admin.Service, sessionService cmd.SessionService) cmd.Guard {
	return &DeveloperGuard{service, sessionService}
}

func (g *DeveloperGuard) Check(ctx *ext.Context, _ string, stdCtx context.Context) (bool, string) {
	userID := ctx.EffectiveSender.Id()

	chatID, err := cmd.GetChatID(g.sessionService, ctx, stdCtx)
	if err != nil {
		return false, "Не удалось определить чат"
	}

	isDev, _ := g.service.IsDeveloper(stdCtx, chatID, userID)
	if !isDev {
		return false, "Эта команда доступна только разработчикам бота"
	}

	return true, ""
}

type DevCreatorGuard struct {
	service        *admin.Service
	sessionService cmd.SessionService
}

func NewDevCreatorGuard(service *admin.Service, sessionService cmd.SessionService) cmd.Guard {
	return &DevCreatorGuard{service, sessionService}
}

func (g *DevCreatorGuard) Check(ctx *ext.Context, _ string, stdCtx context.Context) (bool, string) {
	userID := ctx.EffectiveSender.Id()

	chatID, err := cmd.GetChatID(g.sessionService, ctx, stdCtx)
	if err != nil {
		return false, "Не удалось определить чат"
	}

	role, _ := g.service.GetDevRole(stdCtx, chatID, userID)
	if role != admin.DevRoleCreator {
		return false, ""
	}

	return true, ""
}
