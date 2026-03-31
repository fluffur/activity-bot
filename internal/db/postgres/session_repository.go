package postgres

import (
	"activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/model"
	"activity-bot/internal/session"
	"context"
)

type SessionRepository struct {
	queries *db.Queries
}

var _ session.Repository = (*SessionRepository)(nil)

func NewSessionRepository(queries *db.Queries) *SessionRepository {
	return &SessionRepository{queries}
}

func (r *SessionRepository) SetSession(ctx context.Context, userID int64, chatID int64) error {
	return r.queries.UpsertPMSession(ctx, db.UpsertPMSessionParams{
		UserID:       userID,
		TargetChatID: chatID,
	})
}

func (r *SessionRepository) GetSession(ctx context.Context, userID int64) (int64, error) {
	return r.queries.GetPMSession(ctx, userID)
}

func (r *SessionRepository) GetChat(ctx context.Context, userID int64) (model.Chat, error) {
	chat, err := r.queries.GetChatPMSession(ctx, userID)
	if err != nil {
		return model.Chat{}, err
	}
	return mapChat(chat.Chat), nil
}

func (r *SessionRepository) DeleteSession(ctx context.Context, userID int64) error {
	return r.queries.DeletePMSession(ctx, userID)
}
