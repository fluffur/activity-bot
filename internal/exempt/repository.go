package exempt

import (
	"activity-bot/internal/model"
	"context"
	"time"
)

type Repository interface {
	GetFromChat(ctx context.Context, chatID int64) ([]model.ExemptMember, error)
	Exempt(ctx context.Context, chatID int64, userID int64, exemptUntil time.Time) error
	Get(ctx context.Context, chatID int64, userID int64) (*time.Time, error)
	Remove(ctx context.Context, chatID int64, userID int64) error
	AddRequest(ctx context.Context, request model.ExemptRequest) error
	ApproveRequest(ctx context.Context, request model.ExemptRequest) error
	ApproveRequestWithTx(ctx context.Context, request model.ExemptRequest) error
	RejectRequest(ctx context.Context, chatID, userID int64, messageID int) error
	GetRequest(ctx context.Context, chatID, userID int64, messageID int) (model.ExemptRequest, error)
}
