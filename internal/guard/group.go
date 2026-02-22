package guard

import (
	"activity-bot/internal/cmd"
	"activity-bot/internal/filter"
	"context"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func OnlyGroups(sessionService interface {
	GetActiveChat(ctx context.Context, userID int64) (int64, error)
}) cmd.Guard {
	return cmd.GuardFunc(func(ctx *ext.Context, _ string, stdCtx context.Context) (bool, string) {
		if filter.OnlyGroups(ctx.EffectiveMessage) {
			return true, ""
		}

		if sessionService != nil {
			targetID, err := sessionService.GetActiveChat(stdCtx, ctx.EffectiveUser.Id)
			if err == nil && targetID != 0 {
				return true, ""
			}
		}

		return false, "Команда доступна только в группах или при активной сессии управления чатом в ЛС"
	})
}
