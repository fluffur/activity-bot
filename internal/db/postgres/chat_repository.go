package postgres

import (
	"activity-bot/internal/chat"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/model"
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type ChatRepository struct {
	queries *db.Queries
}

func NewChatRepository(queries *db.Queries) chat.Repository {
	return &ChatRepository{queries}
}

func mapChat(c db.EnsureChatExistsRow) model.Chat {
	return model.Chat{
		ID:                  c.ID,
		NormWarn:            c.NormWarn,
		NormBan:             c.NormBan.Int32,
		NewbieThresholdDays: c.NewbieThresholdDays,
		AISystemPrompt:      c.AiSystemPrompt.String,
		MaxLadder:           c.MaxLadder,
		WelcomeCallMessage:  c.WelcomeCallMessage.String,
		CallOnJoin:          c.CallOnJoin,
		WeekStartDay:        c.WeekStartDay,
		CommandPrefix:       c.CommandPrefix.String,
		AllowPrefixless:     c.AllowPrefixless,
	}
}

func (r *ChatRepository) Ensure(ctx context.Context, c model.Chat) (model.Chat, error) {
	ch, err := r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{
		ID:                  c.ID,
		NormWarn:            c.NormWarn,
		NewbieThresholdDays: 3,
	})
	if err != nil {
		return model.Chat{}, err
	}

	return mapChat(ch), nil
}

func (r *ChatRepository) SetWarnNorm(ctx context.Context, chatID int64, norm int32) error {
	return r.queries.UpdateChatWarnNorm(ctx, db.UpdateChatWarnNormParams{
		NormWarn: norm,
		ID:       chatID,
	})
}

func (r *ChatRepository) SetBanNorm(ctx context.Context, chatID int64, norm int32) error {
	return r.queries.UpdateChatBanNorm(ctx, db.UpdateChatBanNormParams{
		NormBan: pgtype.Int4{Int32: norm, Valid: true},
		ID:      chatID,
	})
}
func (r *ChatRepository) SetNewbieThreshold(ctx context.Context, chatID int64, threshold int32) error {
	return r.queries.UpdateChatNewbieThreshold(ctx, db.UpdateChatNewbieThresholdParams{
		NewbieThresholdDays: threshold,
		ID:                  chatID,
	})
}

func (r *ChatRepository) GetNewbieThreshold(ctx context.Context, chatID int64) (int, error) {
	c, err := r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{
		ID:                  chatID,
		NormWarn:            100,
		NewbieThresholdDays: 3,
	})
	if err != nil {
		return 0, err
	}
	return int(c.NewbieThresholdDays), nil
}

func (r *ChatRepository) GetNorm(ctx context.Context, chatID int64, fallbackNorm int32) (int, error) {
	c, err := r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{
		ID:       chatID,
		NormWarn: fallbackNorm,
	})
	if err != nil {
		return 0, err
	}
	return int(c.NormWarn), nil
}

func (r *ChatRepository) GetChat(ctx context.Context, chatID int64) (model.Chat, error) {
	c, err := r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{
		ID: chatID,
	})
	if err != nil {
		return model.Chat{}, err
	}
	return mapChat(c), nil
}

func (r *ChatRepository) SetChatPrompt(ctx context.Context, chatID int64, prompt string) error {
	return r.queries.SetChatAISystemPrompt(ctx, db.SetChatAISystemPromptParams{
		AiSystemPrompt: pgtype.Text{
			String: prompt,
			Valid:  true,
		},
		ChatID: chatID,
	})
}

func (r *ChatRepository) SetMaxLadder(ctx context.Context, chatID int64, maxLadder int32) error {
	return r.queries.SetChatMaxLadder(ctx, db.SetChatMaxLadderParams{
		ChatID:    chatID,
		MaxLadder: maxLadder,
	})
}

func (r *ChatRepository) SetWelcomeCallMessage(ctx context.Context, chatID int64, message string) error {
	return r.queries.SetChatWelcomeCallMessage(ctx, db.SetChatWelcomeCallMessageParams{
		WelcomeCallMessage: pgtype.Text{
			String: message,
			Valid:  message != "",
		},
		ChatID: chatID,
	})
}

func (r *ChatRepository) UpdateCallOnJoin(ctx context.Context, chatID int64, isEnabled bool) error {
	return r.queries.UpdateChatCallOnJoin(ctx, db.UpdateChatCallOnJoinParams{
		CallOnJoin: isEnabled,
		ChatID:     chatID,
	})
}

func (r *ChatRepository) SetWeekStartDay(ctx context.Context, chatID int64, day int) error {
	return r.queries.UpdateChatWeekStartDay(ctx, db.UpdateChatWeekStartDayParams{
		ChatID:       chatID,
		WeekStartDay: int16(day),
	})
}

func (r *ChatRepository) SetCommandPrefix(ctx context.Context, chatID int64, prefix string) error {
	return r.queries.UpdateChatCommandPrefix(ctx, db.UpdateChatCommandPrefixParams{
		ChatID: chatID,
		CommandPrefix: pgtype.Text{
			String: prefix,
			Valid:  true,
		},
	})
}

func (r *ChatRepository) SetAllowPrefixless(ctx context.Context, chatID int64, allow bool) error {
	return r.queries.UpdateChatAllowPrefixless(ctx, db.UpdateChatAllowPrefixlessParams{
		ChatID:          chatID,
		AllowPrefixless: allow,
	})
}
