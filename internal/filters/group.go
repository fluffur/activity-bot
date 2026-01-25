package filters

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
)

func OnlyGroupsText(msg *gotgbot.Message) bool {
	return msg.Chat.Type != "private" && msg.Text != ""
}
