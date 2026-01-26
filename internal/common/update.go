package common

import (
	"activity-bot/internal/model"
)

type ChatUpdaterService interface {
	UpdateChatMembers(chatID int64, members []model.ChatMemberUpdate) error
}

type ChatAdminsProvider interface {
	GetChatAdmins(chatID int64) ([]model.ChatMemberUpdate, error)
}

type ChatUpdater struct {
	adminsProvider ChatAdminsProvider
	chatUpdater    ChatUpdaterService
}

func NewChatUpdater(adminsProvider ChatAdminsProvider, chatUpdater ChatUpdaterService) *ChatUpdater {
	return &ChatUpdater{adminsProvider, chatUpdater}
}

func (c *ChatUpdater) UpdateChatMembers(chatID int64) (int, error) {
	members, err := c.adminsProvider.GetChatAdmins(chatID)
	if err != nil {
		return 0, err
	}

	if err := c.chatUpdater.UpdateChatMembers(chatID, members); err != nil {
		return 0, err
	}

	return len(members), nil
}
