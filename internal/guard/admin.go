package guard

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type AdminGuard struct {
	service        *admin.Service
	sessionService interface {
		GetActiveChat(ctx context.Context, userID int64) (int64, error)
	}
}

func NewAdminGuard(service *admin.Service, sessionService interface {
	GetActiveChat(ctx context.Context, userID int64) (int64, error)
}) cmd.Guard {
	return &AdminGuard{service, sessionService}
}

func (g *AdminGuard) Check(ctx *ext.Context, _ string, stdCtx context.Context) (bool, string) {
	chatID := ctx.EffectiveChat.Id
	if ctx.EffectiveChat.Type == "private" && g.sessionService != nil {
		targetID, err := g.sessionService.GetActiveChat(stdCtx, ctx.EffectiveSender.Id())
		if err == nil && targetID != 0 {
			chatID = targetID
		}
	}

	if !g.service.CheckIsAdmin(stdCtx, chatID, ctx.EffectiveSender.Id()) {
		return false, ""
	}
	return true, ""
}
