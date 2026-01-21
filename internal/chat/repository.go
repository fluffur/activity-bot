package chat

import (
	"activity-bot/internal/model"
	"context"
	"time"
)

type Repository interface {
	EnsureExists(ctx context.Context, c model.Chat) error
	EnsureMemberExists(ctx context.Context, chatID int64, userID int64) error
	UpsertChatMembers(ctx context.Context, chatID int64, userIDs []int64) error
	GetOrCreate(ctx context.Context, c model.Chat) (model.Chat, error)
	SetNorm(ctx context.Context, chatID int64, norm int) error
	GetChatExemptUsers(ctx context.Context, chatID int64) ([]model.ExemptMember, error)
	GetWeeklyReport(ctx context.Context, chatID int64) ([]model.WeeklyMessageReportMember, error)
	ExemptMember(ctx context.Context, chatID int64, userID int64, exemptUntil time.Time) error
	GetMember(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error)
	RemoveMemberExempt(ctx context.Context, chatID int64, userID int64) error
	AddExemptRequest(ctx context.Context, request model.ExemptRequest) error
	ApproveExemptRequest(ctx context.Context, chatID, userID int64, messageID int) error
	ApproveExemptWithTx(ctx context.Context, request model.ExemptRequest) error
	RejectExemptRequest(ctx context.Context, chatID, userID int64, messageID int) error
	GetExemptRequest(ctx context.Context, chatID, userID int64, messageID int) (model.ExemptRequest, error)

	AddAdmin(ctx context.Context, chatID int64, userID int64) error
	RemoveAdmin(ctx context.Context, chatID int64, userID int64) error
	GetAdmins(ctx context.Context, chatID int64) ([]model.ChatAdmin, error)
	IsAdmin(ctx context.Context, chatID int64, userID int64) (bool, error)
}
