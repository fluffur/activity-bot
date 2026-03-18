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

func (r *RestRepository) GetRestMembers(ctx context.Context, chatID int64) ([]model.ChatMember, error) {
	rows, err := r.queries.GetRestMembers(ctx, chatID)
	if err != nil {
		return nil, err
	}

	members := make([]model.ChatMember, len(rows))
	for i, row := range rows {
		members[i] = mapChatMemberFull(row.ChatMember, row.User)
	}

	return members, nil
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
		MessageID: pgtype.Int8{
			Int64: request.MessageID,
			Valid: request.MessageID != 0,
		},
		Reason: pgtype.Text{
			String: request.Reason,
			Valid:  request.Reason != "",
		},
		Status: db.RestStatus(request.Status),
	})

}

func (r *RestRepository) ApproveRequest(ctx context.Context, request model.RestRequest) error {
	return r.queries.ApproveRestRequest(ctx, db.ApproveRestRequestParams{
		ChatID: request.ChatID,
		UserID: request.UserID,
		MessageID: pgtype.Int8{
			Int64: request.MessageID,
			Valid: request.MessageID != 0,
		},
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
			ChatID: request.ChatID,
			UserID: request.UserID,
			MessageID: pgtype.Int8{
				Int64: request.MessageID,
				Valid: request.MessageID != 0,
			},
		}); err != nil {
			return err
		}

		return nil
	})
}

func (r *RestRepository) RejectRequest(ctx context.Context, chatID, userID, messageID int64) error {
	return r.queries.RejectRestRequest(ctx, db.RejectRestRequestParams{
		ChatID: chatID,
		UserID: userID,
		MessageID: pgtype.Int8{
			Int64: messageID,
			Valid: messageID != 0,
		},
	})

}

func (r *RestRepository) GetRequest(ctx context.Context, chatID, userID, messageID int64) (model.RestRequest, error) {
	er, err := r.queries.GetRestRequest(ctx, db.GetRestRequestParams{
		ChatID: chatID,
		UserID: userID,
		MessageID: pgtype.Int8{
			Int64: messageID,
			Valid: messageID != 0,
		},
	})
	if err != nil {
		return model.RestRequest{}, err
	}

	return mapRestRequest(er), nil

}

func (r *RestRepository) SetRestWithHistory(ctx context.Context, chatID int64, userID int64, messageID int64, until time.Time, reason string) error {
	return r.withTx(ctx, func(q *db.Queries) error {
		if err := q.SetMemberRest(ctx, db.SetMemberRestParams{
			ChatID: chatID,
			UserID: userID,
			RestUntil: pgtype.Timestamptz{
				Time:  until,
				Valid: true,
			},
			RestReason: pgtype.Text{
				String: reason,
				Valid:  reason != "",
			},
		}); err != nil {
			return err
		}

		if err := q.AddRestRequest(ctx, db.AddRestRequestParams{
			ChatID: chatID,
			UserID: userID,
			RestUntil: pgtype.Timestamptz{
				Time:  until,
				Valid: true,
			},
			MessageID: pgtype.Int8{
				Int64: messageID,
				Valid: messageID != 0,
			},
			Status: db.RestStatusApproved,
			Reason: pgtype.Text{
				String: reason,
				Valid:  reason != "",
			},
		}); err != nil {
			return err
		}

		return nil
	})
}

func (r *RestRepository) GetUserRestRequests(ctx context.Context, chatID, userID int64) ([]model.ApprovedRestRequest, error) {
	rows, err := r.queries.GetUserRestRequests(ctx, db.GetUserRestRequestsParams{
		UserID: userID,
		ChatID: chatID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]model.ApprovedRestRequest, len(rows))
	for i, row := range rows {
		result[i] = mapApprovedRestRequest(row.RestRequest, row.ChatMember, row.User)
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

func (r *RestRepository) DeleteRestRequest(ctx context.Context, requestID int64) error {
	return r.queries.DeleteRestRequest(ctx, pgtype.Int8{Int64: requestID, Valid: requestID != 0})
}

func (r *RestRepository) DeleteRestRequestAndEndRest(ctx context.Context, chatID, userID, requestID int64) error {
	return r.withTx(ctx, func(q *db.Queries) error {

		if err := q.EndMemberRest(ctx, db.EndMemberRestParams{
			ChatID: chatID,
			UserID: userID,
		}); err != nil {
			return err
		}

		if err := q.DeleteRestRequest(ctx, pgtype.Int8{Int64: requestID, Valid: requestID != 0}); err != nil {
			return err
		}

		return nil
	})
}
