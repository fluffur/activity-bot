package member

import (
	"activity-bot/internal/model"
	"context"
	"errors"
)

var ErrMemberNotFound = errors.New("member not found")
var ErrInvalidCustomTitle = errors.New("invalid custom title")

type Repository interface {
	GetCustomTitle(ctx context.Context, chatID int64, userID int64) (string, error)
	UpdateCustomTitle(ctx context.Context, chatID int64, userID int64, title *string) error
	UpdateStatus(ctx context.Context, chatID int64, userID int64, role string) error
	FindByChatID(ctx context.Context, chatID int64) ([]model.ChatMember, error)
	GetWithCustomTitles(ctx context.Context, chatID int64) ([]model.ChatMember, error)
	UpsertChatMembers(ctx context.Context, chatID int64, users []model.ChatMemberUpdate) error
	MarkLeftNotInList(ctx context.Context, chatID int64, userIDs []int64) error
	Get(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error)
	Remove(ctx context.Context, chatID int64, userID int64) error
	EnsureExists(ctx context.Context, chatID int64, userID int64, role string) (model.ChatMember, error)
	EnsureFull(ctx context.Context, chatID int64, userID int64, role string, firstName, lastName string, username string, weeklyNorm int32) (model.ChatMember, error)
	SetOnlyNewbies(ctx context.Context, chatID int64, users []*model.User) error
	SetNewbies(ctx context.Context, chatID int64, users []*model.User) error
}
