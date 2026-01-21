package base

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type Handler struct {
}

func (h *Handler) AnswerMessage(ctx context.Context, b *bot.Bot, update *models.Update, text string) {
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:              update.Message.Chat.ID,
		Text:                text,
		ParseMode:           "HTML",
		DisableNotification: true,
	}); err != nil {
		log.Println("Answer message error", err)
		return
	}
}

func (h *Handler) AnswerCallback(ctx context.Context, b *bot.Bot, update *models.Update, text string) {
	if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            text,
	}); err != nil {
		log.Println("Answer callback error", err)
		return
	}
}

func (h *Handler) EditMessage(ctx context.Context, b *bot.Bot, update *models.Update, text string) {
	var m *models.Message
	if update.Message != nil {
		m = update.Message
	} else {
		m = update.CallbackQuery.Message.Message
	}
	if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:    m.Chat.ID,
		MessageID: m.ID,
		Text:      text,
		ParseMode: "HTML",
	}); err != nil {
		log.Println("Edit Message error", err)
		return
	}
}
