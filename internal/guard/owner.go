package guard

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type DeveloperGuard struct {
	service *admin.Service
}

func NewDeveloperGuard(service *admin.Service) cmd.Guard {
	return &DeveloperGuard{service}
}

func (g *DeveloperGuard) Check(ctx *ext.Context, _ string, stdCtx context.Context) (bool, string) {
	userID := ctx.EffectiveSender.Id()

	isDev, _ := g.service.IsDeveloper(stdCtx, userID)
	if !isDev {
		return false, "Эта команда доступна только разработчикам бота"
	}

	return true, ""
}

type DevCreatorGuard struct {
	service *admin.Service
}

func NewDevCreatorGuard(service *admin.Service) cmd.Guard {
	return &DevCreatorGuard{service}
}

func (g *DevCreatorGuard) Check(ctx *ext.Context, _ string, stdCtx context.Context) (bool, string) {
	userID := ctx.EffectiveSender.Id()

	role, _ := g.service.GetDevRole(stdCtx, userID)
	if role != admin.DevRoleCreator {
		return false, ""
	}

	return true, ""
}
