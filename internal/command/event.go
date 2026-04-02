package command

import "github.com/PaulSonOfLars/gotgbot/v2"

func NewChatTitle(msg *gotgbot.Message) bool {
	return msg.NewChatTitle != ""
}
