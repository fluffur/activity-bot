package user

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, user User) error
	Get(ctx context.Context, id int64) (User, error)
	Update(ctx context.Context, user User) error
	SetEmoji(ctx context.Context, id int64, emoji string) error
	SetGender(ctx context.Context, id int64, gender Gender) error
	SetBirthday(ctx context.Context, id int64, birthday time.Time) error
}
