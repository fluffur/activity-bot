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

func (r *ChatRepository) Create(ctx context.Context, c chat.Chat) error {
	return r.queries.CreateChat(ctx, db.CreateChatParams{
		ID:                  c.ID,
		NewbieThresholdDays: c.NewbieThresholdDays,
		AiSystemPrompt:      text(c.AISystemPrompt),
		WeekStartDay:        c.WeekStartDay,
		MaxWarns:            c.MaxWarns,
		CommandPrefix:       text(c.CommandPrefix),
		AllowPrefixless:     c.AllowPrefixless,
		MentionsPerMessage:  c.MentionsPerMessage,
		MentionTypes:        int32(c.MentionTypes),
		Title:               c.Title,
		TagsEnabled:         c.TagsEnabled,
		WeekStartTime: pgtype.Time{
			Microseconds: c.WeekStartTimeMicros,
			Valid:        true,
		},
		RemovedAt:     timestamptz(c.RemovedAt),
		EmojisEnabled: c.EmojisEnabled,
	})
}

func (r *ChatRepository) Get(ctx context.Context, id int64) (chat.Chat, error) {
	u, err := r.queries.GetChatByID(ctx, id)
	if err != nil {
		return chat.Chat{}, err
	}

	return mapChat(u), nil
}
