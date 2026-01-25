package helpers

import (
	"context"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/go-telegram/bot"
)

type AdminService interface {
	IsAdmin(chatID, userID int64) (bool, error)
}

func CheckOwnerOrAdmin(ctx context.Context, b *bot.Bot, adminService AdminService, chatID, userID int64) bool {
	member, err := b.GetChatMember(ctx, &bot.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Failed to get chat member for owner check: %v", err)
		return false
	}
	if member.Owner != nil {
		return true
	}

	isAdmin, err := adminService.IsAdmin(chatID, userID)
	if err != nil {
		log.Printf("Failed to check db admin: %v", err)
		return false
	}
	return isAdmin
}

func IsSenderAdmin(b *gotgbot.Bot, ctx *ext.Context, adminService AdminService) bool {
	senderMember, err := b.GetChatMember(ctx.EffectiveChat.Id, ctx.EffectiveUser.Id, nil)
	if err != nil {
		log.Println("GetChatMember", err)
		return false
	}
	isAdmin, err := adminService.IsAdmin(ctx.EffectiveChat.Id, ctx.EffectiveUser.Id)
	if err != nil {
		log.Println("IsAdmin", err)

		return false
	}

	return senderMember.GetStatus() == "creator" || isAdmin
}
