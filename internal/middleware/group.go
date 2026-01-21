package middleware

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func OnlyGroups(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		m := update.Message
		if m == nil {
			m = update.CallbackQuery.Message.Message
		}
		if m == nil || m.Chat.Type == "private" {
			return
		}
		next(ctx, b, update)
	}
}

func OnlyGroupsMatch(update *models.Update) bool {
	m := update.Message
	if m == nil {
		m = update.CallbackQuery.Message.Message
	}

	return m != nil &&
		m.Chat.Type != "private"
}
