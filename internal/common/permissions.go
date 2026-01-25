package common

import (
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func IsSenderAdmin(b *gotgbot.Bot, ctx *ext.Context, adminService AdminService) bool {
	return IsUserAdmin(b, adminService, ctx.EffectiveChat.Id, ctx.EffectiveSender.Id())
}

func IsSenderCreator(b *gotgbot.Bot, ctx *ext.Context) bool {
	return IsUserCreator(b, ctx.EffectiveChat.Id, ctx.EffectiveSender.Id())
}

func IsUserAdmin(b *gotgbot.Bot, adminService AdminService, chatID, userID int64) bool {
	if IsUserCreator(b, chatID, userID) {
		return true
	}
	isAdmin, err := adminService.IsAdmin(chatID, userID)
	if err != nil {
		log.Println("IsAdmin", err)

		return false
	}

	return isAdmin
}

func IsUserCreator(b *gotgbot.Bot, chatID, userID int64) bool {
	senderMember, err := b.GetChatMember(chatID, userID, nil)
	if err != nil {
		log.Println("GetChatMember", err)
		return false
	}

	return senderMember.GetStatus() == "creator"
}
