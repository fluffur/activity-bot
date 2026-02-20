package guard

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type ModerationGuard struct {
	service *chat.Service
}

func NewModerationGuard(service *chat.Service) cmd.Guard {
	return &ModerationGuard{service}
}

func (g *ModerationGuard) Check(ctx *ext.Context, _ string, stdCtx context.Context) (bool, string) {
	c, err := g.service.GetChat(stdCtx, ctx.EffectiveChat.Id)
	if err != nil {
		return true, "" // Allow if we can't fetch chat info, or handle as error
	}

	if !c.ModerationEnabled {
		return false, ""
	}

	return true, ""
}
