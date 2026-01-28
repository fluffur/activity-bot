package guard

import (
	"activity-bot/internal/command"
	"errors"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func OnlyGroups() command.Guard {
	return command.GuardFunc(func(b *gotgbot.Bot, ctx *ext.Context) error {
		if ctx.EffectiveChat.Type == "private" {
			return errors.New("groups only")
		}
		return nil

	})
}
