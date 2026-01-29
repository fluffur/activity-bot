package postgres

import (
	"activity-bot/internal/chat"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/model"
	"context"
)

type ChatRepository struct {
	queries *db.Queries
}

func NewChatRepository(queries *db.Queries) chat.Repository {
	return &ChatRepository{queries}
}

func mapChat(c db.EnsureChatExistsRow) model.Chat {
	return model.Chat{
		ID:                  c.ID,
		WeeklyNorm:          c.WeeklyNorm,
		NewbieThresholdDays: c.NewbieThresholdDays,
	}
}

func (r *ChatRepository) Ensure(ctx context.Context, c model.Chat) (model.Chat, error) {
	ch, err := r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{
		ID:                  c.ID,
		WeeklyNorm:          c.WeeklyNorm,
		NewbieThresholdDays: 3,
	})
	if err != nil {
		return model.Chat{}, err
	}

	return mapChat(ch), nil
}

func (r *ChatRepository) SetNorm(ctx context.Context, chatID int64, norm int32) error {
	return r.queries.UpdateChatNorm(ctx, db.UpdateChatNormParams{
		WeeklyNorm: norm,
		ID:         chatID,
	})
}

func (r *ChatRepository) SetNewbieThreshold(ctx context.Context, chatID int64, threshold int32) error {
	return r.queries.UpdateChatNewbieThreshold(ctx, db.UpdateChatNewbieThresholdParams{
		NewbieThresholdDays: threshold,
		ID:                  chatID,
	})
}

func (r *ChatRepository) GetNewbieThreshold(ctx context.Context, chatID int64) (int, error) {
	c, err := r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{
		ID:                  chatID,
		WeeklyNorm:          100,
		NewbieThresholdDays: 3,
	})
	if err != nil {
		return 0, err
	}
	return int(c.NewbieThresholdDays), nil
}

func (r *ChatRepository) GetNorm(ctx context.Context, chatID int64, fallbackNorm int32) (int, error) {
	c, err := r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{
		ID:                  chatID,
		WeeklyNorm:          fallbackNorm,
		NewbieThresholdDays: 3,
	})
	if err != nil {
		return 0, err
	}
	return int(c.WeeklyNorm), nil
}
