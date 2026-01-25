package message

import (
	"activity-bot/internal/chat/member"
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service       *Service
	memberService *member.Service
}

func NewHandler(service *Service, memberService *member.Service) *Handler {
	return &Handler{service, memberService}
}

func (h *Handler) Message(_ *gotgbot.Bot, ctx *ext.Context) error {
	u := ctx.EffectiveSender.User
	if u.IsBot {
		return nil
	}

	if _, err := h.memberService.EnsureMemberExists(ctx.EffectiveChat.Id, u.Id, u.Username, u.FirstName, u.LastName); err != nil {
		log.Println("Error", err)
		return err
	}

	if err := h.service.Save(ctx.EffectiveChat.Id, ctx.EffectiveUser.Id); err != nil {
		log.Println("Error", err)
		return err
	}
	return nil
}
