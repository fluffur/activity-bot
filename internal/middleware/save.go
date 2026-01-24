package middleware

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chat/member"
	"activity-bot/internal/user"
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type EnsureMemberExists struct {
	chatService   *chat.Service
	userService   *user.Service
	memberService *member.Service
}

func NewEnsureMemberExists(chatService *chat.Service, userService *user.Service, memberService *member.Service) *EnsureMemberExists {
	return &EnsureMemberExists{chatService, userService, memberService}
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
		if mes.Chat.Type == "private" {
			next(ctx, b, update)
			return
		}

		if u.IsBot {
			next(ctx, b, update)
			return
		}

		if _, err := m.userService.EnsureUserExists(ctx, u.ID, u.Username, u.FirstName, u.LastName); err != nil {
			log.Println("Failed ensure user exists", err)
			return
		}

		if _, err := m.chatService.EnsureChatExists(ctx, mes.Chat.ID); err != nil {
			log.Println("Failed ensure chat exists", err)
			return
		}

		if _, err := m.memberService.EnsureMemberExists(ctx, mes.Chat.ID, u.ID); err != nil {
			log.Println("Failed ensure chat member exists", err)
			return
		}

		next(ctx, b, update)
	}
}
