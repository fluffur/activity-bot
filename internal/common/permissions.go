package common

import (
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type PermissionChecker struct {
	adminService AdminService
	ownerID      int64
}

func NewPermissionChecker(adminService AdminService, ownerID int64) *PermissionChecker {
	return &PermissionChecker{adminService, ownerID}
}

func (c *PermissionChecker) IsAdmin(b *gotgbot.Bot, chatID, userID int64) bool {
	if c.ownerID == userID {
		return true
	}
	isAdmin, err := c.adminService.IsAdmin(chatID, userID)
	if err == nil {
		return isAdmin
	}

	return fallbackCreator(b, chatID, userID)
}

func (c *PermissionChecker) IsCreator(b *gotgbot.Bot, chatID, userID int64) bool {
	if c.ownerID == userID {
		return true
	}
	isCreator, err := c.adminService.IsCreator(chatID, userID)
	if err == nil {
		return isCreator
	}

	return fallbackCreator(b, chatID, userID)
}

func fallbackCreator(b *gotgbot.Bot, chatID, userID int64) bool {
	senderMember, err := b.GetChatMember(chatID, userID, nil)
	if err != nil {
		log.Println("GetChatMember fallback check failed", err)
		return false
	}

	return senderMember.GetStatus() == "creator"
}
