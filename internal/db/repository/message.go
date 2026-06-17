package repository

import (
	db "activity-bot/internal/db/sqlc"
	"activity-bot/internal/events"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type MessageRepository struct {
	queries *db.Queries
}

func NewMessageRepository(queries *db.Queries) events.Repository {
	return &MessageRepository{queries: queries}
}

func (r *MessageRepository) Save(ctx context.Context, chatID, userID, messageID int64) error {
	_, err := r.queries.CreateMessage(ctx, db.CreateMessageParams{
		ChatID: chatID,
		UserID: userID,
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		MessageID: pgtype.Int8{
			Int64: messageID,
			Valid: messageID != 0,
		},
	})

	return err
}
