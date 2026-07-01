package postgres

import (
	"activity-bot/internal/chatmember"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/message"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type MessageRepository struct {
	queries *db.Queries
}

func NewMessageRepository(queries *db.Queries) message.Repository {
	return &MessageRepository{queries: queries}
}

func (r *MessageRepository) Save(ctx context.Context, chatID, userID, messageID int64) error {
	return r.queries.CreateMessage(ctx, db.CreateMessageParams{
		ChatID:    chatID,
		UserID:    userID,
		CreatedAt: timestamptz(time.Now()),
		MessageID: pgtype.Int8{
			Int64: messageID,
			Valid: messageID != 0,
		},
	})
}

func (r *MessageRepository) GetAuthor(ctx context.Context, chatID, messageID int64) (chatmember.ChatMember, error) {
	cm, err := r.queries.GetMessageAuthor(ctx, db.GetMessageAuthorParams{
		MessageID: pgtype.Int8{
			Int64: messageID,
			Valid: true,
		},
		ChatID: chatID,
	})
	if err != nil {
		return chatmember.ChatMember{}, err
	}

	return mapChatMemberFull(cm.ChatMember, db.Chat{}, cm.User), nil
}
