package adapter

import (
	"activity-bot/internal/model"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type TelegramChatAdminsProvider struct {
	bot *gotgbot.Bot
}

func NewTelegramChatAdminsProvider(bot *gotgbot.Bot) *TelegramChatAdminsProvider {
	return &TelegramChatAdminsProvider{bot: bot}
}

func (p *TelegramChatAdminsProvider) GetChatAdmins(chatID int64) ([]model.ChatMemberUpdate, error) {
	admins, err := p.bot.GetChatAdministrators(chatID, nil)
	if err != nil {
		return nil, err
	}

	result := make([]model.ChatMemberUpdate, 0, len(admins))
	for _, admin := range admins {
		chatUser := admin.MergeChatMember()

		if chatUser.User.IsBot {
			continue
		}

		status := admin.GetStatus()
		var rank int16
		if status == "creator" {
			rank = 5
		}

		result = append(result, model.ChatMemberUpdate{
			User: model.User{
				ID:        chatUser.User.Id,
				FirstName: chatUser.User.FirstName,
				LastName:  chatUser.User.LastName,
				Username:  chatUser.User.Username,
			},
			Tag:    chatUser.CustomTitle,
			Status: rank,
		})
	}

	return result, nil
}
