package filter

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
)

func OnlyGroups(msg *gotgbot.Message) bool {
	return msg.Chat.Type != "private"
}
