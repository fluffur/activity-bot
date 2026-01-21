package postgres

import (
	"activity-bot/internal/chat"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/model"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewChatRepository(queries *db.Queries, pool *pgxpool.Pool) chat.Repository {
	return &ChatRepository{queries, pool}
}

func (r *ChatRepository) EnsureExists(ctx context.Context, c model.Chat) error {
	return r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{
		ID:         c.ID,
		WeeklyNorm: c.WeeklyNorm,
	})
}

func (r *ChatRepository) EnsureMemberExists(ctx context.Context, chatID int64, userID int64) error {
	return r.queries.EnsureChatMemberExists(ctx, db.EnsureChatMemberExistsParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *ChatRepository) GetOrCreate(ctx context.Context, c model.Chat) (model.Chat, error) {
	resultChat, err := r.queries.GetOrCreateChat(ctx, db.GetOrCreateChatParams{
		ID:         c.ID,
		WeeklyNorm: c.WeeklyNorm,
	})
	if err != nil {
		return model.Chat{}, err
	}

	return mapChat(resultChat), nil
}

func (r *ChatRepository) withTx(
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

func (r *ChatRepository) SetNorm(ctx context.Context, chatID int64, norm int) error {
	return r.queries.UpdateChatNorm(ctx, db.UpdateChatNormParams{
		WeeklyNorm: int32(norm),
		ID:         chatID,
	})
}

func (r *ChatRepository) GetChatExemptUsers(ctx context.Context, chatID int64) ([]model.ExemptMember, error) {
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

func (r *ChatRepository) ApproveExemptWithTx(ctx context.Context, request model.ExemptRequest) error {
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

func (r *ChatRepository) GetWeeklyReport(ctx context.Context, chatID int64) ([]model.WeeklyMessageReportMember, error) {
	now := time.Now()
	weekday := int(now.Weekday())
	daysSinceMonday := (weekday + 6) % 7
	monday := now.AddDate(0, 0, -daysSinceMonday)
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())

	rows, err := r.queries.WeeklyMessageReport(ctx, db.WeeklyMessageReportParams{
		ChatID: chatID,
		CreatedAt: pgtype.Timestamptz{
			Time:  monday,
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	result := make([]model.WeeklyMessageReportMember, len(rows))
	for i, row := range rows {
		result[i] = mapWeeklyReportRow(row)
	}

	return result, nil
}

func (r *ChatRepository) ExemptMember(ctx context.Context, chatID int64, userID int64, exemptUntil time.Time) error {
	return r.queries.ExemptChatMember(ctx, db.ExemptChatMemberParams{
		ExemptUntil: pgtype.Timestamptz{
			Time:  exemptUntil,
			Valid: true,
		},
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *ChatRepository) GetMember(ctx context.Context, chatID int64, userID int64) (model.ChatMember, error) {
	m, err := r.queries.GetChatMember(ctx, db.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
	})
	if err != nil {
		return model.ChatMember{}, err
	}

	return mapMember(m), nil
}

func (r *ChatRepository) RemoveMemberExempt(ctx context.Context, chatID int64, userID int64) error {
	return r.queries.RemoveChatMemberExempt(ctx, db.RemoveChatMemberExemptParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *ChatRepository) AddExemptRequest(ctx context.Context, request model.ExemptRequest) error {
	return r.queries.AddExemptRequest(ctx, db.AddExemptRequestParams{
		ChatID: int64(request.ChatID),
		UserID: int64(request.UserID),
		ExemptUntil: pgtype.Timestamptz{
			Time:  request.ExemptUntil,
			Valid: true,
		},
		MessageID: int64(request.MessageID),
	})
}

func (r *ChatRepository) ApproveExemptRequest(ctx context.Context, chatID, userID int64, messageID int) error {
	return r.queries.ApproveExemptRequest(ctx, db.ApproveExemptRequestParams{
		ChatID:    chatID,
		UserID:    userID,
		MessageID: int64(messageID),
	})
}

func (r *ChatRepository) RejectExemptRequest(ctx context.Context, chatID, userID int64, messageID int) error {
	return r.queries.RejectExemptRequest(ctx, db.RejectExemptRequestParams{
		ChatID:    chatID,
		UserID:    userID,
		MessageID: int64(messageID),
	})
}

func (r *ChatRepository) GetExemptRequest(ctx context.Context, chatID, userID int64, messageID int) (model.ExemptRequest, error) {
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

func mapMember(m db.ChatMember) model.ChatMember {
	result := model.ChatMember{
		ChatID: m.ChatID,
		UserID: m.UserID,
	}
	if m.ExemptUntil.Valid {
		result.ExemptUntil = &m.ExemptUntil.Time
	}

	return result
}

func mapChatExemptUsersRow(row db.ChatExemptUsersRow) model.ExemptMember {
	displayName := row.Username.String
	if displayName == "" {
		displayName = row.FirstName.String
		if row.LastName.Valid {
			displayName += " " + row.LastName.String
		}
	}
	return model.ExemptMember{
		UserID:      row.UserID,
		DisplayName: displayName,
		ExemptUntil: row.ExemptUntil.Time,
	}
}

func mapWeeklyReportRow(row db.WeeklyMessageReportRow) model.WeeklyMessageReportMember {
	displayName := row.Username.String
	if displayName == "" {
		displayName = row.FirstName.String
		if row.LastName.Valid {
			displayName += " " + row.LastName.String
		}
	}

	return model.WeeklyMessageReportMember{
		UserID:        row.UserID,
		DisplayName:   displayName,
		MessagesCount: int32(row.MessagesCount),
		WeeklyNorm:    row.WeeklyNorm,
		NormDone:      row.NormDone,
	}
}

func mapChat(c db.Chat) model.Chat {
	return model.Chat{
		ID:         c.ID,
		WeeklyNorm: c.WeeklyNorm,
	}
}

func (r *ChatRepository) AddAdmin(ctx context.Context, chatID int64, userID int64) error {
	return r.queries.AddChatAdmin(ctx, db.AddChatAdminParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *ChatRepository) RemoveAdmin(ctx context.Context, chatID int64, userID int64) error {
	return r.queries.RemoveChatAdmin(ctx, db.RemoveChatAdminParams{
		ChatID: chatID,
		UserID: userID,
	})
}

func (r *ChatRepository) GetAdmins(ctx context.Context, chatID int64) ([]model.ChatAdmin, error) {
	rows, err := r.queries.GetChatAdmins(ctx, chatID)
	if err != nil {
		return nil, err
	}

	admins := make([]model.ChatAdmin, len(rows))
	for i, row := range rows {
		displayName := row.Username.String
		if displayName == "" {
			displayName = row.FirstName.String
			if row.LastName.Valid {
				displayName += " " + row.LastName.String
			}
		}

		admins[i] = model.ChatAdmin{
			UserID:      row.ID,
			DisplayName: displayName,
			CreatedAt:   row.CreatedAt.Time,
		}
	}
	return admins, nil
}

func (r *ChatRepository) IsAdmin(ctx context.Context, chatID int64, userID int64) (bool, error) {
	return r.queries.IsChatAdmin(ctx, db.IsChatAdminParams{
		ChatID: chatID,
		UserID: userID,
	})
}
