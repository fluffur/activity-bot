package match

import "github.com/go-telegram/bot/models"

func LeftMember(update *models.Update) bool {
	return update.Message != nil && update.Message.LeftChatMember != nil && !update.Message.LeftChatMember.IsBot
}
