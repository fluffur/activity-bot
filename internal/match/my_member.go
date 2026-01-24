package match

import "github.com/go-telegram/bot/models"

func PromotedToAdministrator(update *models.Update) bool {
	return update.MyChatMember != nil && update.MyChatMember.NewChatMember.Administrator != nil
}
