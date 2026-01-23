package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/exempt"
	"activity-bot/internal/model"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExemptRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewExemptRepository(queries *db.Queries, pool *pgxpool.Pool) exempt.Repository {
	return &ExemptRepository{queries, pool}
}

func (r *ExemptRepository) GetFromChat(ctx context.Context, chatID int64) ([]model.ExemptMember, error) {
	rows, err := r.queries.ChatExemptUsers(ctx, chatID)
	if err != nil {
		return nil, err
	}

	result := make([]model.ExemptMember, len(rows))
	for i, row := range rows {
		result[i] = mapChatExemptUsersRow(row)
	}

	return result, nil
}

func (r *ExemptRepository) Exempt(ctx context.Context, chatID int64, userID int64, exemptUntil time.Time) error {
	return r.queries.ExemptChatMember(ctx, db.ExemptChatMemberParams{
		ExemptUntil: pgtype.Timestamptz{
			Time:  exemptUntil,
			Valid: true,
		},
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *ExemptRepository) Get(ctx context.Context, chatID int64, userID int64) (*time.Time, error) {
	m, err := r.queries.GetChatMember(ctx, db.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	var exemptUntil *time.Time
	if m.ExemptUntil.Valid {
		exemptUntil = &m.ExemptUntil.Time
	}
	return exemptUntil, nil
}

func (r *ExemptRepository) Remove(ctx context.Context, chatID int64, userID int64) error {
	return r.queries.RemoveChatMemberExempt(ctx, db.RemoveChatMemberExemptParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *ExemptRepository) AddRequest(ctx context.Context, request model.ExemptRequest) error {
	return r.queries.AddExemptRequest(ctx, db.AddExemptRequestParams{
		ChatID: request.ChatID,
		UserID: request.UserID,
		ExemptUntil: pgtype.Timestamptz{
			Time:  request.ExemptUntil,
			Valid: true,
		},
		MessageID: request.MessageID,
	})

}

func (r *ExemptRepository) ApproveRequest(ctx context.Context, request model.ExemptRequest) error {
	return r.queries.ApproveExemptRequest(ctx, db.ApproveExemptRequestParams{
		ChatID:    request.ChatID,
		UserID:    request.UserID,
		MessageID: request.MessageID,
	})

}

func (r *ExemptRepository) ApproveRequestWithTx(ctx context.Context, request model.ExemptRequest) error {
	return r.withTx(ctx, func(q *db.Queries) error {
		if err := q.ExemptChatMember(ctx, db.ExemptChatMemberParams{
			ChatID: int64(request.ChatID),
			UserID: int64(request.UserID),
			ExemptUntil: pgtype.Timestamptz{
				Time:  request.ExemptUntil,
				Valid: true,
			},
		}); err != nil {
			return err
		}

		if err := q.ApproveExemptRequest(ctx, db.ApproveExemptRequestParams{
			ChatID:    int64(request.ChatID),
			UserID:    int64(request.UserID),
			MessageID: int64(request.MessageID),
		}); err != nil {
			return err
		}

		return nil
	})
}

func (r *ExemptRepository) RejectRequest(ctx context.Context, chatID, userID int64, messageID int) error {

	return r.queries.RejectExemptRequest(ctx, db.RejectExemptRequestParams{
		ChatID:    chatID,
		UserID:    userID,
		MessageID: int64(messageID),
	})

}

func (r *ExemptRepository) GetRequest(ctx context.Context, chatID, userID int64, messageID int) (model.ExemptRequest, error) {
	er, err := r.queries.GetExemptRequest(ctx, db.GetExemptRequestParams{
		ChatID:    chatID,
		UserID:    userID,
		MessageID: int64(messageID),
	})
	if err != nil {
		return model.ExemptRequest{}, err
	}

	return mapExemptRequest(er), nil

}

func (r *ExemptRepository) withTx(
	ctx context.Context,
	fn func(q *db.Queries) error,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	q := r.queries.WithTx(tx)

	if err = fn(q); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func mapExemptRequest(er db.ExemptRequest) model.ExemptRequest {
	return model.ExemptRequest{
		ChatID:      er.ChatID,
		UserID:      er.UserID,
		RequestedAt: er.RequestedAt.Time,
		ExemptUntil: er.ExemptUntil.Time,
		Status:      string(er.Status),
		MessageID:   er.MessageID,
	}
}

func mapChatExemptUsersRow(row db.ChatExemptUsersRow) model.ExemptMember {
	fullName := row.FirstName.String
	if row.LastName.Valid {
		fullName += " " + row.LastName.String
	}
	var username *string
	if row.Username.Valid {
		username = &row.Username.String
	}
	return model.ExemptMember{
		User: model.User{
			ID:        row.UserID,
			FirstName: row.FirstName.String,
			LastName:  row.LastName.String,
			Username:  username,
		},
		ExemptUntil: row.ExemptUntil.Time,
	}
}
