package ban

import (
	"context"
	"time"
)

type Repository interface {
	BanUser(ctx context.Context, userID int64, reason string, expiresAt time.Time) error
	UnbanUser(ctx context.Context, userID int64) error
	IsUserBanned(ctx context.Context, userID int64) (bool, error)
	GetUserBan(ctx context.Context, userID int64) (UserBan, error)

	BanChat(ctx context.Context, chatID int64, reason string, expiresAt time.Time) error
	UnbanChat(ctx context.Context, chatID int64) error
	IsChatBanned(ctx context.Context, chatID int64) (bool, error)
	GetChatBan(ctx context.Context, chatID int64) (ChatBan, error)
}
