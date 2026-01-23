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

func (r *ChatRepository) Ensure(ctx context.Context, c model.Chat) (model.Chat, error) {
	ch, err := r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{
		ID:         c.ID,
		WeeklyNorm: c.WeeklyNorm,
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

func (r *ChatRepository) GetNorm(ctx context.Context, chatID int64, fallbackNorm int32) (int, error) {
	c, err := r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{
		ID:         chatID,
		WeeklyNorm: fallbackNorm,
	})
	if err != nil {
		return 0, err
	}
	return int(c.WeeklyNorm), nil
}

func mapChat(c db.Chat) model.Chat {
	return model.Chat{
		ID:         c.ID,
		WeeklyNorm: c.WeeklyNorm,
	}
}
