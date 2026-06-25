package repository

import (
	db "activity-bot/internal/db/sqlc"
	"activity-bot/internal/norm"
	"context"
)

type NormRepository struct {
	queries *db.Queries
}

func NewNormRepository(queries *db.Queries) *NormRepository {
	return &NormRepository{queries: queries}
}

func (r *NormRepository) Get(ctx context.Context, chatID int64, name string) (norm.Norm, error) {
	n, err := r.queries.GetNorm(ctx, db.GetNormParams{
		ChatID: chatID,
		Name:   name,
	})
	if err != nil {
		return norm.Norm{}, err
	}
	return mapNorm(n), nil
}

func (r *NormRepository) Set(ctx context.Context, chatID int64, name string, value int32) error {
	return r.queries.SetNorm(ctx, db.SetNormParams{
		ChatID: chatID,
		Name:   name,
		Value:  value,
	})
}

func (r *NormRepository) List(ctx context.Context, chatID int64) ([]norm.Norm, error) {
	norms, err := r.queries.ListNorms(ctx, chatID)
	if err != nil {
		return nil, err
	}
	return mapList(norms, mapNorm), nil
}

func (r *NormRepository) Delete(ctx context.Context, chatID int64, name string) error {
	return r.queries.DeleteNorm(ctx, db.DeleteNormParams{
		ChatID: chatID,
		Name:   name,
	})
}

func (r *NormRepository) ListWithMembers(
	ctx context.Context,
	chatID int64,
) ([]norm.Norm, error) {
	rows, err := r.queries.ListNormsWithMembers(ctx, chatID)
	if err != nil {
		return nil, err
	}

	normsByID := make(map[int64]*norm.Norm)

	for _, row := range rows {
		n, ok := normsByID[row.ChatNorm.ID]
		if !ok {
			nn := mapNorm(row.ChatNorm)

			normsByID[row.ChatNorm.ID] = &nn
			n = &nn
		}

		if row.UserID.Valid {
			n.UserIDs = append(
				n.UserIDs,
				row.UserID.Int64,
			)
		}
	}

	result := make([]norm.Norm, 0, len(normsByID))

	for _, n := range normsByID {
		result = append(result, *n)
	}

	return result, nil
}
