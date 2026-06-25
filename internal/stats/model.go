package stats

import "activity-bot/internal/chatmember"

type ChatStats struct {
	ChatMember    chatmember.ChatMember
	MessagesCount int64
}
