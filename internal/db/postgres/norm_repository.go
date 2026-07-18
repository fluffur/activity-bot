package postgres

import (
	"activity-bot/internal/chatmember"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/norm"
	"context"
)

type NormRepository struct {
	queries *db.Queries
}

func NewNormRepository(queries *db.Queries) *NormRepository {
	return &NormRepository{queries: queries}
}

func (r *NormRepository) GetByName(ctx context.Context, chatID int64, name string) (norm.Norm, error) {
	n, err := r.queries.GetNorm(ctx, db.GetNormParams{
		ChatID: chatID,
		Name:   name,
	})
	if err != nil {
		return norm.Norm{}, err
	}

	return mapNorm(n), nil
}

func (r *NormRepository) GetNormMembers(ctx context.Context, normID int64) ([]chatmember.ChatMember, error) {
	cms, err := r.queries.GetNormMembers(ctx, normID)
	if err != nil {
		return nil, err
	}

	return mapList(cms, func(t db.GetNormMembersRow) chatmember.ChatMember {
		return mapChatMemberFull(t.ChatMember, db.Chat{}, t.User)
	}), nil
}

func (r *NormRepository) Set(ctx context.Context, chatID int64, name string, value int32) (int64, error) {
	return r.queries.SetNorm(ctx, db.SetNormParams{
		ChatID: chatID,
		Name:   name,
		Value:  value,
	})
}

func (r *NormRepository) Delete(ctx context.Context, normID int64) error {
	return r.queries.DeleteNorm(ctx, normID)
}

func (r *NormRepository) List(
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

func (r *NormRepository) Assign(ctx context.Context, normID int64, userIDs []int64) error {
	return r.queries.AssignNormMembers(ctx, db.AssignNormMembersParams{
		NormID:  normID,
		UserIds: userIDs,
	})
}

func (r *NormRepository) Unassign(ctx context.Context, normID int64, userIDs []int64) error {
	return r.queries.UnassignNormMembers(ctx, db.UnassignNormMembersParams{
		NormID:  normID,
		UserIds: userIDs,
	})
}

func (r *NormRepository) GetByID(ctx context.Context, normID int64) (norm.Norm, error) {
	n, err := r.queries.GetNormByID(ctx, normID)
	if err != nil {
		return norm.Norm{}, err
	}

	return mapNorm(n), nil
}
