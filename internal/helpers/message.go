package helpers

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func SendMessage(ctx context.Context, b *bot.Bot, update *models.Update, text string) {
	isLinkPreviewDisabled := true
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: "HTML",
		LinkPreviewOptions: &models.LinkPreviewOptions{
			IsDisabled: &isLinkPreviewDisabled,
		},
	}); err != nil {

		log.Println("Answer message error", err)
		return
	}
}

func AnswerCallback(ctx context.Context, b *bot.Bot, update *models.Update, text string) {
	if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: update.CallbackQuery.ID,
		Text:            text,
	}); err != nil {
		log.Println("Answer callback error", err)
		return
	}
}

func EditMessage(ctx context.Context, b *bot.Bot, update *models.Update, text string) {
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
