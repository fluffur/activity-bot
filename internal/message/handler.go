package message

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service}
}

func (h *Handler) Message(ctx context.Context, b *bot.Bot, update *models.Update) {
	log.Println(update.Message.Text)
	if err := h.service.Save(ctx, update.Message.Chat.ID, update.Message.From); err != nil {
		log.Println("Error", err)
		return
	}
}
