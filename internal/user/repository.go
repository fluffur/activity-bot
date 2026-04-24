package user

import (
	"activity-bot/internal/model"
	"context"
)

type Repository interface {
	Ensure(ctx context.Context, id int64, username, firstName, lastName string, isBot bool) (model.User, error)
	GetUser(ctx context.Context, id int64) (model.User, error)
	GetByUsername(ctx context.Context, username string) (model.User, error)
	UpsertUsers(ctx context.Context, users []model.User) error
	SetGender(ctx context.Context, userID int64, gender string) error
	SetEmoji(ctx context.Context, userID int64, emojis model.Emojis) error
}
