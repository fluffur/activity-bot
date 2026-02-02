package guard

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type AdminGuard struct {
	service *admin.Service
}

func NewAdminGuard(service *admin.Service) cmd.Guard {
	return &AdminGuard{service}
}

func (g *AdminGuard) Check(ctx *ext.Context, _ string) (bool, string) {
	if !g.service.CheckIsAdmin(ctx.EffectiveChat.Id, ctx.EffectiveSender.Id()) {
		return false, "Только создатель и администраторы могут выполнить эту команду"
	}
	return true, ""
}
