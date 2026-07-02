package rest

import (
	"activity-bot/internal/chatmember"
	"context"
	"time"
)

type Request struct {
	ID          int64
	ChatID      int64
	UserID      int64
	RequestedAt time.Time
	RestUntil   time.Time
	UpdatedAt   time.Time
	Status      string
	MessageID   int64
	Reason      string

	ChatMember chatmember.ChatMember
}

type Repository interface {
	GetRestMembers(ctx context.Context, chatID int64) ([]chatmember.ChatMember, error)
	SetRest(ctx context.Context, chatID int64, userID int64, until time.Time, reason string) error
	EndMemberRest(ctx context.Context, chatID int64, userID int64) error
	AddRequest(ctx context.Context, request Request) error
	ApproveRequest(ctx context.Context, request Request) error
	ApproveRequestWithTx(ctx context.Context, request Request) error
	RejectRequest(ctx context.Context, chatID, userID, messageID int64) error
	GetRequest(ctx context.Context, chatID, userID, messageID int64) (Request, error)
	SetRestWithHistory(ctx context.Context, chatID int64, userID int64, messageID int64, until time.Time, reason string) error
	GetUserRestRequests(ctx context.Context, chatID, userID int64) ([]Request, error)
	DeleteRestRequest(ctx context.Context, requestID int64) error
	DeleteRestRequestAndEndRest(ctx context.Context, chatID, userID, requestID int64) error
}
