package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/reward"
	"context"
)

type RewardRepository struct {
	queries *db.Queries
}

func NewRewardRepository(queries *db.Queries) *RewardRepository {
	return &RewardRepository{queries: queries}
}

func (r *RewardRepository) AddReward(ctx context.Context, rw reward.Reward) error {
	return r.queries.AddReward(ctx, db.AddRewardParams{
		ChatID:   rw.ChatID,
		UserID:   rw.UserID,
		Rank:     rw.Rank,
		Reason:   rw.Reason,
		AuthorID: rw.AuthorID,
	})
}

func (r *RewardRepository) RemoveReward(ctx context.Context, rewardID int64) error {
	return r.queries.RemoveReward(ctx, rewardID)
}

func (r *RewardRepository) ListRewards(ctx context.Context, chatID, userID int64) ([]reward.Reward, error) {
	rws, err := r.queries.ListUserRewards(ctx, db.ListUserRewardsParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	return mapList(rws, mapReward), nil
}

func mapReward(rw db.Reward) reward.Reward {
	return reward.Reward{
		ID:        rw.ID,
		ChatID:    rw.ChatID,
		UserID:    rw.UserID,
		Rank:      rw.Rank,
		Reason:    rw.Reason,
		CreatedAt: rw.CreatedAt.Time,
	}
}
