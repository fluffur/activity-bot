package chat

import "activity-bot/internal/model"

type ChatMemberUpdate struct {
	User        model.User
	CustomTitle string
}
