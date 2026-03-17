package user

import (
	"activity-bot/internal/model"
	"context"
)

type Repository interface {
	Ensure(ctx context.Context, id int64, username, firstName, lastName string) (model.User, error)
	Get(ctx context.Context, id int64) (model.User, error)
	GetByUsername(ctx context.Context, username string) (model.User, error)
	UpsertUsers(ctx context.Context, users []model.User) error
	GetByTag(ctx context.Context, chatID int64, tag string) ([]model.ChatMember, error)
	SetGender(ctx context.Context, userID int64, gender string) error
	SetEmoji(ctx context.Context, userID int64, emoji string) error
	SetCustomEmojiID(ctx context.Context, userID int64, emojiID string) error
}
