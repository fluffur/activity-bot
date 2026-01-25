package message

import (
	"log"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service}
}

func (h *Handler) Message(b *gotgbot.Bot, ctx *ext.Context) error {
	if err := h.service.Save(ctx.EffectiveChat.Id, ctx.EffectiveUser.Id); err != nil {
		log.Println("Error", err)
		return err
	}
	return nil
}
