package events

import "context"

type Repository interface {
	Save(ctx context.Context, chatID, userID, messageID int64) error
}
