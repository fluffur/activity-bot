package postgres

import (
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/model"
	"activity-bot/internal/rest"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RestRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewRestRepository(queries *db.Queries, pool *pgxpool.Pool) rest.Repository {
	return &RestRepository{queries, pool}
}

func (r *RestRepository) GetFromChat(ctx context.Context, chatID int64) ([]model.RestMember, error) {
	rows, err := r.queries.GetRestMembers(ctx, chatID)
	if err != nil {
		return nil, err
	}

	result := make([]model.RestMember, len(rows))
	for i, row := range rows {
		result[i] = mapChatRestUsersRow(row)
	}

	return result, nil
}

func (r *RestRepository) SetRest(ctx context.Context, chatID int64, userID int64, restUntil time.Time, reason string) error {
	return r.queries.SetMemberRest(ctx, db.SetMemberRestParams{
		RestUntil: pgtype.Timestamptz{
			Time:  restUntil,
			Valid: true,
		},
		ChatID: chatID,
		UserID: userID,
		RestReason: pgtype.Text{
			String: reason,
			Valid:  reason != "",
		},
	})
}

func (r *RestRepository) EndMemberRest(ctx context.Context, chatID int64, userID int64) error {
	return r.queries.EndMemberRest(ctx, db.EndMemberRestParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *RestRepository) AddRequest(ctx context.Context, request model.RestRequest) error {
	return r.queries.AddRestRequest(ctx, db.AddRestRequestParams{
		ChatID: request.ChatID,
		UserID: request.UserID,
		RestUntil: pgtype.Timestamptz{
			Time:  request.RestUntil,
			Valid: true,
		},
		MessageID: request.MessageID,
		Reason: pgtype.Text{
			String: request.Reason,
			Valid:  request.Reason != "",
		},
	})

}

func (r *RestRepository) ApproveRequest(ctx context.Context, request model.RestRequest) error {
	return r.queries.ApproveRestRequest(ctx, db.ApproveRestRequestParams{
		ChatID:    request.ChatID,
		UserID:    request.UserID,
		MessageID: request.MessageID,
	})

}

func (r *RestRepository) ApproveRequestWithTx(ctx context.Context, request model.RestRequest) error {
	return r.withTx(ctx, func(q *db.Queries) error {
		if err := q.SetMemberRest(ctx, db.SetMemberRestParams{
			ChatID: request.ChatID,
			UserID: request.UserID,
			RestUntil: pgtype.Timestamptz{
				Time:  request.RestUntil,
				Valid: true,
			},
			RestReason: pgtype.Text{
				String: request.Reason,
				Valid:  request.Reason != "",
			},
		}); err != nil {
			return err
		}

		if err := q.ApproveRestRequest(ctx, db.ApproveRestRequestParams{
			ChatID:    request.ChatID,
			UserID:    request.UserID,
			MessageID: request.MessageID,
		}); err != nil {
			return err
		}

		return nil
	})
}

func (r *RestRepository) RejectRequest(ctx context.Context, chatID, userID, messageID int64) error {
	return r.queries.RejectRestRequest(ctx, db.RejectRestRequestParams{
		ChatID:    chatID,
		UserID:    userID,
		MessageID: messageID,
	})

}

func (r *RestRepository) GetRequest(ctx context.Context, chatID, userID, messageID int64) (model.RestRequest, error) {
	er, err := r.queries.GetRestRequest(ctx, db.GetRestRequestParams{
		ChatID:    chatID,
		UserID:    userID,
		MessageID: messageID,
	})
	if err != nil {
		return model.RestRequest{}, err
	}

	return mapRestRequest(er), nil

}

func (r *RestRepository) GetAllActiveRests(ctx context.Context) ([]model.RestExpirePayload, error) {
	rows, err := r.queries.GetAllActiveRests(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.RestExpirePayload, len(rows))
	for i, row := range rows {
		result[i] = model.RestExpirePayload{
			ChatID:     row.ChatID,
			UserID:     row.UserID,
			RestUntil:  row.RestUntil.Time,
			RestReason: row.RestReason.String,
		}
	}

	return result, nil
}

func (r *RestRepository) withTx(
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

func mapRestRequest(er db.RestRequest) model.RestRequest {
	return model.RestRequest{
		ChatID:      er.ChatID,
		UserID:      er.UserID,
		RequestedAt: er.RequestedAt.Time,
		RestUntil:   er.RestUntil.Time,
		Status:      string(er.Status),
		MessageID:   er.MessageID,
	}
}

func mapChatRestUsersRow(row db.GetRestMembersRow) model.RestMember {
	fullName := row.FirstName.String
	if row.LastName.Valid {
		fullName += " " + row.LastName.String
	}
	var username *string
	if row.Username.Valid {
		username = &row.Username.String
	}
	return model.RestMember{
		User: model.User{
			ID:        row.UserID,
			FirstName: row.FirstName.String,
			LastName:  row.LastName.String,
			Username:  username,
		},
		RestUntil:   row.RestUntil.Time,
		Status:      row.Status,
		CustomTitle: row.CustomTitle.String,
	}
}
