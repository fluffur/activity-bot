package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/message"
	"activity-bot/internal/model"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type MessageRepository struct {
	queries *db.Queries
}

func NewMessageRepository(queries *db.Queries) message.Repository {
	return &MessageRepository{queries}
}

func (r *MessageRepository) Save(ctx context.Context, m model.Message) error {
	if _, err := r.queries.CreateMessage(ctx, db.CreateMessageParams{
		MessageID: pgtype.Int8{
			Int64: m.MessageID,
			Valid: true,
		},
		ChatID: m.ChatID,
		UserID: m.UserID,
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
	}); err != nil {
		return err
	}

	return nil
}
