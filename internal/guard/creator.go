package guard

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type CreatorGuard struct {
	service *admin.Service
}

func NewCreatorGuard(service *admin.Service) cmd.Guard {
	return &CreatorGuard{service}
}

func (g *CreatorGuard) Check(ctx *ext.Context, _ string, stdCtx context.Context) (bool, string) {
	if !g.service.CheckIsCreator(stdCtx, ctx.EffectiveChat.Id, ctx.EffectiveSender.Id()) {
		return false, ""
	}
	return true, ""
}
