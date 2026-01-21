package middleware

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/model"
	"activity-bot/internal/user"
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type EnsureMemberExists struct {
	chatRepo          chat.Repository
	userRepo          user.Repository
	defaultWeeklyNorm int32
}

func NewEnsureMemberExists(chatRepo chat.Repository, userRepo user.Repository, defaultWeeklyNorm int32) *EnsureMemberExists {
	return &EnsureMemberExists{
		chatRepo:          chatRepo,
		userRepo:          userRepo,
		defaultWeeklyNorm: defaultWeeklyNorm,
	}
}

func (m *EnsureMemberExists) Handle(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		var mes *models.Message
		var u *models.User
		if update.Message != nil {
			mes = update.Message
			u = mes.From
		} else {
			mes = update.CallbackQuery.Message.Message
			u = &update.CallbackQuery.From
		}

		if err := m.userRepo.EnsureExists(ctx, u.ID, u.Username, u.FirstName, u.LastName); err != nil {
			log.Println("Failed ensure user exists", err)
			return
		}

		if err := m.chatRepo.EnsureExists(ctx, model.NewChat(mes.Chat.ID, m.defaultWeeklyNorm)); err != nil {
			log.Println("Failed ensure chat exists", err)
			return
		}

		if err := m.chatRepo.EnsureMemberExists(ctx, mes.Chat.ID, u.ID); err != nil {
			log.Println("Failed ensure chat member exists", err)
			return
		}

		next(ctx, b, update)
	}
}
