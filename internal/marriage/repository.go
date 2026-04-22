package marriage

import "context"

type Repository interface {
	CreateMarriageRequest(ctx context.Context, chatID, fromUserID, toUserID int64) (MarriageRequest, error)
	GetActiveMarriage(ctx context.Context, chatID, userID int64) (*Marriage, error)
	GetMarriageBetweenUsers(ctx context.Context, chatID, user1ID, user2ID int64) (*Marriage, error)
	GetActiveMarriageRequest(ctx context.Context, chatID, fromUserID, toUserID int64) (*MarriageRequest, error)
	UpdateMarriageRequestStatus(ctx context.Context, requestID, chatID int64, status RequestStatus) error
	CreateMarriage(ctx context.Context, chatID, user1ID, user2ID int64) (Marriage, error)
	DivorceMarriage(ctx context.Context, marriageID, chatID int64) error
	ListActiveMarriages(ctx context.Context, chatID int64) ([]Marriage, error)
}
