package common

import (
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func IsSenderAdmin(b *gotgbot.Bot, ctx *ext.Context, adminService AdminService) bool {
	if IsSenderCreator(b, ctx) {
		return true
	}
	isAdmin, err := adminService.IsAdmin(ctx.EffectiveChat.Id, ctx.EffectiveUser.Id)
	if err != nil {
		log.Println("IsAdmin", err)

		return false
	}

	return isAdmin
}

func IsSenderCreator(b *gotgbot.Bot, ctx *ext.Context) bool {
	senderMember, err := b.GetChatMember(ctx.EffectiveChat.Id, ctx.EffectiveUser.Id, nil)
	if err != nil {
		log.Println("GetChatMember", err)
		return false
	}

	return senderMember.GetStatus() == "creator"
}
