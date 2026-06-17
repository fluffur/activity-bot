package repository

import (
	"activity-bot/internal/chat"
	db "activity-bot/internal/db/sqlc"
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type ChatRepository struct {
	queries *db.Queries
}

func NewChatRepository(queries *db.Queries) chat.Repository {
	return &ChatRepository{queries: queries}
}

func (r *ChatRepository) Create(ctx context.Context, chat chat.Chat) error {
	return r.queries.CreateChat(ctx, db.CreateChatParams{
		ID:                  chat.ID,
		NewbieThresholdDays: chat.NewbieThresholdDays,
		AiSystemPrompt:      text(chat.AISystemPrompt),
		WeekStartDay:        chat.WeekStartDay,
		MaxWarns:            chat.MaxWarns,
		CommandPrefix:       text(chat.CommandPrefix),
		AllowPrefixless:     chat.AllowPrefixless,
		MentionsPerMessage:  chat.MentionsPerMessage,
		MentionTypes:        chat.MentionTypes,
		Title:               chat.Title,
		TagsEnabled:         chat.TagsEnabled,
		WeekStartTime: pgtype.Time{
			Microseconds: chat.WeekStartTime,
			Valid:        true,
		},
		RemovedAt:     timestamptz(chat.RemovedAt),
		EmojisEnabled: chat.EmojisEnabled,
	})
}

func (r *ChatRepository) GetByID(ctx context.Context, id int64) (chat.Chat, error) {
	u, err := r.queries.GetChatByID(ctx, id)
	if err != nil {
		return chat.Chat{}, err
	}
	return mapChat(u), nil
}
