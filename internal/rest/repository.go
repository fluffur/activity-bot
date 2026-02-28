package rest

import (
	"activity-bot/internal/model"
	"context"
	"time"
)

type Repository interface {
	GetFromChat(ctx context.Context, chatID int64) ([]model.RestMember, error)
	SetRest(ctx context.Context, chatID int64, userID int64, until time.Time) error
	GetRestUntil(ctx context.Context, chatID int64, userID int64) (*time.Time, error)
	EndMemberRest(ctx context.Context, chatID int64, userID int64) error
	AddRequest(ctx context.Context, request model.RestRequest) error
	ApproveRequest(ctx context.Context, request model.RestRequest) error
	ApproveRequestWithTx(ctx context.Context, request model.RestRequest) error
	RejectRequest(ctx context.Context, chatID, userID, messageID int64) error
	GetRequest(ctx context.Context, chatID, userID, messageID int64) (model.RestRequest, error)
	GetAllActiveRests(ctx context.Context) ([]model.RestExpirePayload, error)
}
