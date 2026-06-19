package pmsession

import (
	"activity-bot/internal/chat"
	"context"
)

type Repository interface {
	GetChat(ctx context.Context, userID int64) (chat.Chat, error)
}
