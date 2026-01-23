package helpers

import (
	"activity-bot/internal/model"
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type ChatMemberService interface {
	UpdateChatMembers(ctx context.Context, chatID int64, members []model.ChatMemberUpdate) error
}

func UpdateChatMembers(ctx context.Context, b *bot.Bot, service ChatMemberService, chatID int64) (int, error) {
	admins, err := b.GetChatAdministrators(ctx, &bot.GetChatAdministratorsParams{
		ChatID: chatID,
	})
	if err != nil {
		return 0, err
	}

	members := make([]model.ChatMemberUpdate, 0, len(admins))
	for _, admin := range admins {
		var chatUser *models.User
		var customTitle string

		if admin.Administrator != nil {
			chatUser = &admin.Administrator.User
			customTitle = admin.Administrator.CustomTitle
		} else if admin.Owner != nil {
			chatUser = admin.Owner.User
			customTitle = admin.Owner.CustomTitle
		} else {
			chatUser = nil
		}
		if chatUser == nil {
			continue
		}

		if chatUser.IsBot {
			continue
		}
		members = append(members, model.ChatMemberUpdate{
			User: model.User{
				ID:        chatUser.ID,
				FirstName: chatUser.FirstName,
				LastName:  chatUser.LastName,
				Username:  &chatUser.Username,
			},
			CustomTitle: customTitle,
		})
	}

	if err := service.UpdateChatMembers(ctx, chatID, members); err != nil {
		return 0, err
	}
	return len(members), nil
}
