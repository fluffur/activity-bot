package guard

import (
	"activity-bot/internal/command"

	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func OnlyGroups() command.Guard {
	return command.GuardFunc(func(ctx *ext.Context) (bool, string) {
		if ctx.EffectiveChat.Type == "private" {
			return false, "Команда доступна только в группах"
		}
		return true, ""

	})
}
