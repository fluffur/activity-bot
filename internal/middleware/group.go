package middleware

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func OnlyGroups(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		m := update.Message
		if m == nil && update.CallbackQuery != nil {
			m = update.CallbackQuery.Message.Message
		} else {
			next(ctx, b, update)
			return
		}
		if m == nil || m.Chat.Type == "private" {
			return
		}
		next(ctx, b, update)
	}
}
