package guard

import (
	"activity-bot/internal/admin"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type AdminGuard struct {
	service *admin.Service
}

func NewAdminGuard(service *admin.Service) *AdminGuard {
	return &AdminGuard{service}
}

func (g *AdminGuard) Check(b *gotgbot.Bot, ctx *ext.Context) error {
	if !g.service.CheckIsAdmin(ctx.EffectiveChat.Id, ctx.EffectiveSender.Id()) {
		_, err := ctx.EffectiveMessage.Reply(b, "Только создатель и администраторы могут выполнить эту команду", nil)
		return err
	}
	return nil
}
