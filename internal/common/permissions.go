package common

import (
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func IsSenderAdmin(b *gotgbot.Bot, ctx *ext.Context, adminService AdminService) bool {
	return IsUserAdmin(b, adminService, ctx.EffectiveChat.Id, ctx.EffectiveSender.Id())
}

func IsSenderCreator(b *gotgbot.Bot, ctx *ext.Context, adminService AdminService) bool {
	return IsUserCreator(b, adminService, ctx.EffectiveChat.Id, ctx.EffectiveSender.Id())
}

func IsUserAdmin(b *gotgbot.Bot, adminService AdminService, chatID, userID int64) bool {
	if IsUserCreator(b, adminService, chatID, userID) {
		return true
	}

	isAdmin, err := adminService.IsAdmin(chatID, userID)
	if err != nil {
		log.Println("IsBotAdmin check failed", err)
		return false
	}

	return isAdmin
}

func IsUserCreator(b *gotgbot.Bot, adminService AdminService, chatID, userID int64) bool {
	role, err := adminService.GetRole(chatID, userID)
	if err == nil {
		return role == "creator"
	}

	senderMember, err := b.GetChatMember(chatID, userID, nil)
	if err != nil {
		log.Println("GetChatMember fallback check failed", err)
		return false
	}

	return senderMember.GetStatus() == "creator"
}
