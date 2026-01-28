package guard

import (
	"activity-bot/internal/command"
	"activity-bot/internal/filter"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func OnlyGroups() command.Guard {
	return command.GuardFunc(func(ctx *ext.Context) (bool, string) {
		if !filter.OnlyGroups(ctx.EffectiveMessage) {
			return false, "Команда доступна только в группах"
		}
		return true, ""

	})
}
