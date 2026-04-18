package postgres

import (
	"activity-bot/internal/chat"
	db "activity-bot/internal/db/postgres/sqlc"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ChatRepository struct {
	queries *db.Queries
}

func NewChatRepository(queries *db.Queries) chat.Repository {
	return &ChatRepository{queries}
}

func (r *ChatRepository) SetTagsEnabled(ctx context.Context, chatID int64, enabled bool) error {
	return r.queries.SetChatTagsEnabled(ctx, db.SetChatTagsEnabledParams{
		TagsEnabled: enabled,
		ID:          chatID,
	})
}

func (r *ChatRepository) Ensure(ctx context.Context, c model.Chat) (model.Chat, error) {
	ch, err := r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{
		ID:    c.ID,
		Title: c.Title,
	})
	if err != nil {
		return model.Chat{}, err
	}

	return mapChatFromRow(ch), nil
}

func (r *ChatRepository) SetTitle(ctx context.Context, chatID int64, title string) error {
	return r.queries.UpdateChatTitle(ctx, db.UpdateChatTitleParams{
		Title: title,
		ID:    chatID,
	})
}

func (r *ChatRepository) SetWarnNorm(ctx context.Context, chatID int64, norm int32) error {
	return r.queries.UpdateChatWarnNorm(ctx, db.UpdateChatWarnNormParams{
		ID:       chatID,
		NormWarn: pgtype.Int4{Int32: norm, Valid: norm != 0},
	})
}

func (r *ChatRepository) SetBanNorm(ctx context.Context, chatID int64, norm int32) error {
	return r.queries.UpdateChatBanNorm(ctx, db.UpdateChatBanNormParams{
		NormBan: pgtype.Int4{Int32: norm, Valid: norm != 0},
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
		ID: chatID,
	})
	if err != nil {
		return 0, err
	}
	return int(c.NewbieThresholdDays), nil
}

func (r *ChatRepository) GetChat(ctx context.Context, chatID int64) (model.Chat, error) {
	c, err := r.queries.GetChat(ctx, chatID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {

			c, err := r.queries.EnsureChatExists(ctx, db.EnsureChatExistsParams{ID: chatID})
			if err != nil {
				return model.Chat{}, err
			}
			return mapChatFromRow(c), nil

		}
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

func (r *ChatRepository) SetMentionsPerMessage(ctx context.Context, chatID int64, count int32) error {
	return r.queries.UpdateChatMentionsPerMessage(ctx, db.UpdateChatMentionsPerMessageParams{
		ChatID:             chatID,
		MentionsPerMessage: count,
	})
}

func (r *ChatRepository) SetMentionTypes(ctx context.Context, chatID int64, types int32) error {
	return r.queries.UpdateChatMentionTypes(ctx, db.UpdateChatMentionTypesParams{
		ChatID:       chatID,
		MentionTypes: types,
	})
}

func (r *ChatRepository) GetChatsWithoutNorm(ctx context.Context, userID int64) ([]model.ChatWithoutNorm, error) {
	chats, err := r.queries.GetAllUserChatsWithoutNorm(ctx, userID)
	if err != nil {
		return nil, err
	}

	return mapChatsWithoutNorm(chats), nil

}

func mapChatsWithoutNorm(chats []db.GetAllUserChatsWithoutNormRow) []model.ChatWithoutNorm {
	result := make([]model.ChatWithoutNorm, len(chats))
	for i, c := range chats {
		result[i] = model.ChatWithoutNorm{
			ID:        c.ID,
			Title:     c.Title,
			NormBan:   c.NormBan.Int32,
			NormWarn:  c.NormWarn.Int32,
			WeekCount: c.WeekCount,
		}
	}
	return result
}

func (r *ChatRepository) SetWeekStartTime(ctx context.Context, chatID int64, time string) error {
	return r.queries.UpdateChatWeekStartTime(ctx, db.UpdateChatWeekStartTimeParams{
		ChatID:        chatID,
		WeekStartTime: pgtype.Time{Microseconds: helpers.TimeToMicroseconds(time), Valid: true},
	})
}

func (r *ChatRepository) GetChatsWithoutTitle(ctx context.Context) ([]model.Chat, error) {
	chats, err := r.queries.GetChatsWithoutTitle(ctx)
	if err != nil {
		return nil, err
	}
	return mapMany(chats, mapChat), nil
}

func (r *ChatRepository) GetUserManagedChats(ctx context.Context, userID int64) ([]model.Chat, error) {
	chats, err := r.queries.GetUserManagedChats(ctx, userID)
	if err != nil {
		return nil, err
	}

	mapped := make([]model.Chat, len(chats))
	for i, c := range chats {
		mapped[i] = mapChatFromRow(db.EnsureChatExistsRow(c))
	}

	return mapped, nil
}

func (r *ChatRepository) GetAllChats(ctx context.Context) ([]model.Chat, error) {
	chats, err := r.queries.GetAllChats(ctx)
	if err != nil {
		return nil, err
	}
	mapped := make([]model.Chat, len(chats))
	for i, c := range chats {
		mapped[i] = mapChatFromRow(db.EnsureChatExistsRow(c))
	}

	return mapped, nil
}

func (r *ChatRepository) GetChatsWithEnabledBroadcast(ctx context.Context) ([]model.Chat, error) {
	chats, err := r.queries.GetChatsWithEnabledBroadcast(ctx)
	if err != nil {
		return nil, err
	}
	return mapMany(chats, mapChat), nil
}

func (r *ChatRepository) SetChatBroadcast(ctx context.Context, chatID int64, enabled bool) error {
	return r.queries.SetChatBroadcast(ctx, db.SetChatBroadcastParams{
		BroadcastEnabled: enabled,
		ID:               chatID,
	})
}

func (r *ChatRepository) GetCommandPermissions(ctx context.Context, chatID int64) (map[string]model.Status, error) {
	perms, err := r.queries.GetCommandPermissions(ctx, chatID)
	if err != nil {
		return nil, err
	}
	res := make(map[string]model.Status)
	for _, p := range perms {
		res[p.CommandKey] = model.Status(p.RequiredStatus)
	}
	return res, nil
}

func (r *ChatRepository) GetCommandPermission(ctx context.Context, chatID int64, key string) (model.Status, error) {
	p, err := r.queries.GetCommandPermission(ctx, db.GetCommandPermissionParams{
		ChatID:     chatID,
		CommandKey: key,
	})
	if err != nil {
		return 0, err
	}
	return model.Status(p.RequiredStatus), nil
}

func (r *ChatRepository) SetCommandPermission(ctx context.Context, chatID int64, key string, status model.Status) error {
	return r.queries.SetCommandPermission(ctx, db.SetCommandPermissionParams{
		ChatID:         chatID,
		CommandKey:     key,
		RequiredStatus: int16(status),
	})
}

func (r *ChatRepository) Remove(ctx context.Context, chatID int64) error {
	return r.queries.RemoveChat(ctx, db.RemoveChatParams{
		RemovedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		ID: chatID,
	})
}
