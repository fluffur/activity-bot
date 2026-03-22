package guard

import (
	"activity-bot/internal/admin"
	"activity-bot/internal/cmd"
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type DeveloperGuard struct {
	service     *admin.Service
	developerID int64
}

func NewDeveloperGuard(service *admin.Service, developerID int64) cmd.Guard {
	return &DeveloperGuard{service, developerID}
}

func (g *DeveloperGuard) Check(ctx *ext.Context, _ string, _ context.Context) (bool, string) {
	return g.developerID == ctx.EffectiveSender.Id(), "s"
}
