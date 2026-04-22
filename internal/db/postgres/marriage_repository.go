package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/marriage"
	"activity-bot/internal/model"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type MarriageRepository struct {
	queries *db.Queries
}

func NewMarriageRepository(queries *db.Queries) marriage.Repository {
	return &MarriageRepository{queries: queries}
}

func (r *MarriageRepository) CreateMarriageRequest(ctx context.Context, chatID, fromUserID, toUserID int64) (marriage.MarriageRequest, error) {
	req, err := r.queries.CreateMarriageRequest(ctx, db.CreateMarriageRequestParams{
		ChatID:     chatID,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
	})
	if err != nil {
		return marriage.MarriageRequest{}, err
	}
	return mapMarriageRequest(req), nil
}

func (r *MarriageRepository) GetActiveMarriage(ctx context.Context, chatID, userID int64) (*marriage.Marriage, error) {
	m, err := r.queries.GetActiveMarriage(ctx, db.GetActiveMarriageParams{
		ChatID:  chatID,
		User1ID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	mapped := mapActiveMarriageRow(m)
	return &mapped, nil
}

func (r *MarriageRepository) GetMarriageBetweenUsers(ctx context.Context, chatID, user1ID, user2ID int64) (*marriage.Marriage, error) {
	m, err := r.queries.GetMarriageBetweenUsers(ctx, db.GetMarriageBetweenUsersParams{
		ChatID:    chatID,
		User1ID:   user1ID,
		User1ID_2: user2ID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	mapped := mapMarriageBetweenUsersRow(m)
	return &mapped, nil
}

func (r *MarriageRepository) GetActiveMarriageRequest(ctx context.Context, chatID, fromUserID, toUserID int64) (*marriage.MarriageRequest, error) {
	req, err := r.queries.GetActiveMarriageRequest(ctx, db.GetActiveMarriageRequestParams{
		ChatID:     chatID,
		FromUserID: fromUserID,
		ToUserID:   toUserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	mapped := mapMarriageRequest(req)
	return &mapped, nil
}

func (r *MarriageRepository) UpdateMarriageRequestStatus(ctx context.Context, requestID, chatID int64, status marriage.RequestStatus) error {
	return r.queries.UpdateMarriageRequestStatus(ctx, db.UpdateMarriageRequestStatusParams{
		ID:     requestID,
		ChatID: chatID,
		Status: db.MarriageRequestStatus(status),
	})
}

func (r *MarriageRepository) CreateMarriage(ctx context.Context, chatID, user1ID, user2ID int64) (marriage.Marriage, error) {
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}

	m, err := r.queries.CreateMarriage(ctx, db.CreateMarriageParams{
		ChatID:  chatID,
		User1ID: user1ID,
		User2ID: user2ID,
	})
	if err != nil {
		return marriage.Marriage{}, err
	}
	return mapCreatedMarriage(m), nil
}

func (r *MarriageRepository) DivorceMarriage(ctx context.Context, marriageID, chatID int64) error {
	return r.queries.DivorceMarriage(ctx, db.DivorceMarriageParams{
		ID:     marriageID,
		ChatID: chatID,
	})
}

func (r *MarriageRepository) ListActiveMarriages(ctx context.Context, chatID int64) ([]marriage.Marriage, error) {
	rows, err := r.queries.ListActiveMarriages(ctx, chatID)
	if err != nil {
		return nil, err
	}
	return mapMany(rows, mapListActiveMarriagesRow), nil
}

func mapCreatedMarriage(m db.Marriage) marriage.Marriage {
	return marriage.Marriage{
		ID:        m.ID,
		ChatID:    m.ChatID,
		MarriedAt: m.MarriedAt.Time,
		User1: model.ChatMember{
			ChatID: m.ChatID,
			User: model.User{
				ID: m.User1ID,
			},
		},
		User2: model.ChatMember{
			ChatID: m.ChatID,
			User: model.User{
				ID: m.User2ID,
			},
		},
	}
}

func mapActiveMarriageRow(row db.GetActiveMarriageRow) marriage.Marriage {
	return marriage.Marriage{
		ID:        row.Marriage.ID,
		ChatID:    row.Marriage.ChatID,
		MarriedAt: row.Marriage.MarriedAt.Time,
		User1:     mapChatMemberFull(row.ChatMember, row.User),
		User2:     mapChatMemberFull(row.ChatMember_2, row.User_2),
	}
}

func mapMarriageBetweenUsersRow(row db.GetMarriageBetweenUsersRow) marriage.Marriage {
	return marriage.Marriage{
		ID:        row.Marriage.ID,
		ChatID:    row.Marriage.ChatID,
		MarriedAt: row.Marriage.MarriedAt.Time,
		User1:     mapChatMemberFull(row.ChatMember, row.User),
		User2:     mapChatMemberFull(row.ChatMember_2, row.User_2),
	}
}

func mapListActiveMarriagesRow(row db.ListActiveMarriagesRow) marriage.Marriage {
	return marriage.Marriage{
		ID:        row.Marriage.ID,
		ChatID:    row.Marriage.ChatID,
		MarriedAt: row.Marriage.MarriedAt.Time,
		User1:     mapChatMemberFull(row.ChatMember, row.User),
		User2:     mapChatMemberFull(row.ChatMember_2, row.User_2),
	}
}

func mapMarriageRequest(r db.MarriageRequest) marriage.MarriageRequest {
	return marriage.MarriageRequest{
		ID:         r.ID,
		ChatID:     r.ChatID,
		FromUserID: r.FromUserID,
		ToUserID:   r.ToUserID,
		Status:     marriage.RequestStatus(r.Status),
	}
}
