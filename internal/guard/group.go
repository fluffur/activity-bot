package guard

import (
	"activity-bot/internal/cmd"
	"activity-bot/internal/filter"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func OnlyGroups() cmd.Guard {
	return cmd.GuardFunc(func(ctx *ext.Context, _ string) (bool, string) {
		if !filter.OnlyGroups(ctx.EffectiveMessage) {
			return false, "Команда доступна только в группах"
		}
		return true, ""

	})
}
