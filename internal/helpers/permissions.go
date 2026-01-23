package helpers

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
)

type AdminService interface {
	IsAdmin(ctx context.Context, chatID, userID int64) (bool, error)
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

	isAdmin, err := adminService.IsAdmin(ctx, chatID, userID)
	if err != nil {
		log.Printf("Failed to check db admin: %v", err)
		return false
	}
	return isAdmin
}
