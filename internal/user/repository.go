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
	GetByCustomTitle(ctx context.Context, chatID int64, title string) (model.User, error)
}
