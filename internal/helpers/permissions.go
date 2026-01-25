package helpers

import (
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type AdminService interface {
	IsAdmin(chatID, userID int64) (bool, error)
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
