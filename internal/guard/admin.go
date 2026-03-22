package guard

import (
	"activity-bot/internal/cmd"
	"activity-bot/internal/member"
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type SessionService interface {
	GetActiveChat(ctx context.Context, userID int64) (int64, error)
}

type AdminGuard struct {
	service        *member.Service
	sessionService SessionService
	status         int16
}

func NewStatusGuard(service *member.Service, sessionService SessionService, status int16) cmd.Guard {
	return &AdminGuard{service, sessionService, status}
}

func (g AdminGuard) Check(ctx *ext.Context, _ string, stdCtx context.Context) (bool, string) {
	chatID := ctx.EffectiveChat.Id
	if ctx.EffectiveChat.Type == "private" && g.sessionService != nil {
		targetID, err := g.sessionService.GetActiveChat(stdCtx, ctx.EffectiveSender.Id())
		if err == nil && targetID != 0 {
			chatID = targetID
		}
	}

	m, err := g.service.GetChatMember(stdCtx, chatID, ctx.EffectiveSender.Id())
	if err != nil {
		return false, ""
	}
	return m.Status >= g.status, ""
}
