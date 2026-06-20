package repository

import (
	"activity-bot/internal/chat"
	db "activity-bot/internal/db/sqlc"
	"activity-bot/internal/pmsession"
	"context"
)

type PMSessionRepository struct {
	queries *db.Queries
}

func NewPMSessionRepository(queries *db.Queries) pmsession.Repository {
	return &PMSessionRepository{queries: queries}
}

func (r *PMSessionRepository) GetChat(ctx context.Context, userID int64) (chat.Chat, error) {
	s, err := r.queries.GetChatPMSession(ctx, userID)
	if err != nil {
		return chat.Chat{}, err
	}

	return mapChat(s.Chat), nil
}
