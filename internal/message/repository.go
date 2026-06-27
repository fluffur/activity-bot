package message

import (
	"activity-bot/internal/chatmember"
	"context"
)

type Repository interface {
	Save(ctx context.Context, chatID, userID, messageID int64) error
	GetAuthor(ctx context.Context, chatID, messageID int64) (chatmember.ChatMember, error)
}
