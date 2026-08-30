package postgres

import (
	"activity-bot/internal/ban"
	db "activity-bot/internal/db/postgres/sqlc"
	"context"
	"time"
)

type BanRepository struct {
	queries *db.Queries
}

func NewBanRepository(queries *db.Queries) ban.Repository {
	return &BanRepository{queries: queries}
}

func (r *BanRepository) BanUser(
	ctx context.Context,
	userID int64,
	reason string,
	expiresAt time.Time,
) error {
	return r.queries.BanUser(ctx, db.BanUserParams{
		UserID:    userID,
		Reason:    text(reason),
		ExpiresAt: timestamptz(expiresAt),
	})
}

func (r *BanRepository) UnbanUser(ctx context.Context, userID int64) error {
	return r.queries.UnbanUser(ctx, userID)
}

func (r *BanRepository) IsUserBanned(ctx context.Context, userID int64) (bool, error) {
	return r.queries.IsUserBanned(ctx, userID)
}

func (r *BanRepository) GetUserBan(ctx context.Context, userID int64) (ban.UserBan, error) {
	b, err := r.queries.GetUserBan(ctx, userID)
	if err != nil {
		return ban.UserBan{}, err
	}

	return ban.UserBan{
		UserID:    b.UserID,
		Reason:    b.Reason.String,
		CreatedAt: b.CreatedAt.Time,
		ExpiresAt: b.ExpiresAt.Time,
	}, nil
}

func (r *BanRepository) BanChat(
	ctx context.Context,
	chatID int64,
	reason string,
	expiresAt time.Time,
) error {
	return r.queries.BanChat(ctx, db.BanChatParams{
		ChatID:    chatID,
		Reason:    text(reason),
		ExpiresAt: timestamptz(expiresAt),
	})
}

func (r *BanRepository) UnbanChat(ctx context.Context, chatID int64) error {
	return r.queries.UnbanChat(ctx, chatID)
}

func (r *BanRepository) IsChatBanned(ctx context.Context, chatID int64) (bool, error) {
	return r.queries.IsChatBanned(ctx, chatID)
}

func (r *BanRepository) GetChatBan(ctx context.Context, chatID int64) (ban.ChatBan, error) {
	b, err := r.queries.GetChatBan(ctx, chatID)
	if err != nil {
		return ban.ChatBan{}, err
	}

	return ban.ChatBan{
		ChatID:    b.ChatID,
		Reason:    b.Reason.String,
		CreatedAt: b.CreatedAt.Time,
		ExpiresAt: b.ExpiresAt.Time,
	}, nil
}
