package guard

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type CreatorGuard struct {
	service        *admin.Service
	sessionService interface {
		GetActiveChat(ctx context.Context, userID int64) (int64, error)
	}
}

func NewCreatorGuard(service *admin.Service, sessionService interface {
	GetActiveChat(ctx context.Context, userID int64) (int64, error)
}) cmd.Guard {
	return &CreatorGuard{service, sessionService}
}

func (g *CreatorGuard) Check(ctx *ext.Context, _ string, stdCtx context.Context) (bool, string) {
	chatID := ctx.EffectiveChat.Id
	if ctx.EffectiveChat.Type == "private" && g.sessionService != nil {
		targetID, err := g.sessionService.GetActiveChat(stdCtx, ctx.EffectiveSender.Id())
		if err == nil && targetID != 0 {
			chatID = targetID
		}
	}

	if !g.service.CheckIsCreator(stdCtx, chatID, ctx.EffectiveSender.Id()) {
		return false, ""
	}
	return true, ""
}
