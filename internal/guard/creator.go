package guard

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/command"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type CreatorGuard struct {
	service *admin.Service
}

func NewCreatorGuard(service *admin.Service) command.Guard {
	return &CreatorGuard{service}
}

func (g *CreatorGuard) Check(ctx *ext.Context) (bool, string) {
	if !g.service.CheckIsCreator(ctx.EffectiveChat.Id, ctx.EffectiveSender.Id()) {
		return false, "Только создатель может выполнить эту команду"
	}
	return true, ""
}
