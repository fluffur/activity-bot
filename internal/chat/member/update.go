package member

import (
	"activity-bot/internal/model"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type ChatMemberService interface {
	UpdateChatMembers(chatID int64, members []model.ChatMemberUpdate) error
}

func UpdateChatMembers(b *gotgbot.Bot, service ChatMemberService, chatID int64) (int, error) {
	admins, err := b.GetChatAdministrators(chatID, nil)
	if err != nil {
		return 0, err
	}

	members := make([]model.ChatMemberUpdate, 0, len(admins))
	for _, admin := range admins {
		chatUser := admin.MergeChatMember()

		if chatUser.User.IsBot {
			continue
		}

		members = append(members, model.ChatMemberUpdate{
			User: model.User{
				ID:        chatUser.User.Id,
				FirstName: chatUser.User.FirstName,
				LastName:  chatUser.User.LastName,
				Username:  &chatUser.User.Username,
			},
			CustomTitle: chatUser.CustomTitle,
			Role:        admin.GetStatus(),
		})
	}

	if err := service.UpdateChatMembers(chatID, members); err != nil {
		return 0, err
	}
	return len(members), nil
}
