package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/norm"
	"context"
)

type NormRepository struct {
	queries *db.Queries
}

func NewNormRepository(queries *db.Queries) norm.Repository {
	return &NormRepository{queries}
}

func (r *NormRepository) Create(ctx context.Context, chatID int64, name string, value int32) error {
	return r.queries.CreateNorm(ctx, db.CreateNormParams{
		ChatID: chatID,
		Name:   name,
		Value:  value,
	})
}

func (r *NormRepository) Get(ctx context.Context, chatID int64, name string) (norm.ChatNorm, error) {
	n, err := r.queries.GetNorm(ctx, db.GetNormParams{
		ChatID: chatID,
		Name:   name,
	})
	if err != nil {
		return norm.ChatNorm{}, err
	}
	return mapChatNorm(n), nil
}

func (r *NormRepository) ListAll(ctx context.Context, chatID int64) ([]norm.ChatNorm, error) {
	norms, err := r.queries.ListNorms(ctx, chatID)
	if err != nil {
		return nil, err
	}

	return mapMany(norms, mapChatNorm), nil
}
