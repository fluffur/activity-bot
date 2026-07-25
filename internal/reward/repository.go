package reward

import "context"

type Repository interface {
	AddReward(ctx context.Context, rw Reward) error
	RemoveReward(ctx context.Context, rewardID int64) error
	ListRewards(ctx context.Context, chatID, userID int64) ([]Reward, error)
}
