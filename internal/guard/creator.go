package guard

import (
	"activity-bot/internal/admin"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type CreatorGuard struct {
	service *admin.Service
}

func NewCreatorGuard(service *admin.Service) *CreatorGuard {
	return &CreatorGuard{service}
}

func (g *CreatorGuard) Check(b *gotgbot.Bot, ctx *ext.Context) error {
	if !g.service.CheckIsCreator(ctx.EffectiveChat.Id, ctx.EffectiveSender.Id()) {
		_, err := ctx.EffectiveMessage.Reply(b, "Только создатель может выполнить эту команду", nil)
		return err
	}
	return nil
}
