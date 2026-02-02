package postgres

import (
	"activity-bot/internal/chat"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/model"
	"context"

	"github.com/jackc/pgx/v5/pgtype"
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
		GeminiSystemPrompt:  c.GeminiSystemPrompt.String,
		MaxLadder:           c.MaxLadder,
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
		ID:         chatID,
		WeeklyNorm: fallbackNorm,
	})
	if err != nil {
		return 0, err
	}
	return int(c.WeeklyNorm), nil
}

func (r *ChatRepository) GetChat(ctx context.Context, chatID int64) (model.Chat, error) {
	c, err := r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{
		ID: chatID,
	})
	if err != nil {
		return model.Chat{}, err
	}
	return mapChat(c), nil
}

func (r *ChatRepository) SetChatPrompt(ctx context.Context, chatID int64, prompt string) error {
	return r.queries.SetChatGeminiSystemPrompt(ctx, db.SetChatGeminiSystemPromptParams{
		GeminiSystemPrompt: pgtype.Text{
			String: prompt,
			Valid:  true,
		},
		ChatID: chatID,
	})
}

func (r *ChatRepository) SetMaxLadder(ctx context.Context, chatID int64, maxLadder int32) error {
	return r.queries.SetChatMaxLadder(ctx, db.SetChatMaxLadderParams{
		ChatID:    chatID,
		MaxLadder: maxLadder,
	})
}
