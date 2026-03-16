package rest

import (
	"activity-bot/internal/model"
	"context"
	"time"
)

type Repository interface {
	GetFromChat(ctx context.Context, chatID int64) ([]model.RestMember, error)
	SetRest(ctx context.Context, chatID int64, userID int64, until time.Time, reason string) error
	EndMemberRest(ctx context.Context, chatID int64, userID int64) error
	AddRequest(ctx context.Context, request model.RestRequest) error
	ApproveRequest(ctx context.Context, request model.RestRequest) error
	ApproveRequestWithTx(ctx context.Context, request model.RestRequest) error
	RejectRequest(ctx context.Context, chatID, userID, messageID int64) error
	GetRequest(ctx context.Context, chatID, userID, messageID int64) (model.RestRequest, error)
	GetAllActiveRests(ctx context.Context) ([]model.RestExpirePayload, error)
	GetApprovedRequests(ctx context.Context) ([]model.ApprovedRestRequest, error)
	GetUserApprovedRequests(ctx context.Context, userID int64) ([]model.ApprovedRestRequest, error)
	SetRestWithHistory(ctx context.Context, chatID int64, userID int64, messageID int64, until time.Time, reason string) error
	GetUserRestRequests(ctx context.Context, userID int64) ([]model.ApprovedRestRequest, error)
}
