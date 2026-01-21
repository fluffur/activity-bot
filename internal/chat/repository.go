package chat

import (
	"activity-bot/internal/model"
	"context"
	"time"
)

type Repository interface {
	EnsureExists(ctx context.Context, c model.Chat) error
	EnsureMemberExists(ctx context.Context, chatID int64, userID int64) error
	GetOrCreate(ctx context.Context, c model.Chat) (model.Chat, error)
	SetNorm(ctx context.Context, chatID int64, norm int) error
	GetChatExemptUsers(ctx context.Context, chatID int64) ([]model.ExemptUsersRow, error)
	GetWeeklyReport(ctx context.Context, chatID int64) ([]model.WeeklyMessageReportRow, error)
	ExemptMember(ctx context.Context, chatID int64, userID int64, exemptUntil time.Time) error
	GetMember(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error)
	RemoveMemberExempt(ctx context.Context, chatID int64, userID int64) error
	AddExemptRequest(ctx context.Context, request model.ExemptRequest) error
	ApproveExemptRequest(ctx context.Context, chatID, userID int64, messageID int) error
	ApproveExemptWithTx(ctx context.Context, request model.ExemptRequest) error
	RejectExemptRequest(ctx context.Context, chatID, userID int64, messageID int) error
	GetExemptRequest(ctx context.Context, chatID, userID int64, messageID int) (model.ExemptRequest, error)
}
