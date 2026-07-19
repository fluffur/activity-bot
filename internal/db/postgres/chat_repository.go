package postgres

import (
	"activity-bot/internal/chat"
	db "activity-bot/internal/db/postgres/sqlc"
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type ChatRepository struct {
	queries *db.Queries
}

func NewChatRepository(queries *db.Queries) chat.Repository {
	return &ChatRepository{queries: queries}
}

func (r *ChatRepository) Create(ctx context.Context, c chat.Chat) error {
	return r.queries.CreateChat(ctx, db.CreateChatParams{
		ID:                  c.ID,
		NewbieThresholdDays: c.NewbieThresholdDays,
		AiSystemPrompt:      text(c.AISystemPrompt),
		WeekStartDay:        c.WeekStartDay,
		MaxWarns:            c.MaxWarns,
		CommandPrefix:       text(c.CommandPrefix),
		AllowPrefixless:     c.AllowPrefixless,
		MentionsPerMessage:  c.MentionsPerMessage,
		MentionTypes:        int32(c.MentionTypes),
		Title:               c.Title,
		TagsEnabled:         c.TagsEnabled,
		WeekStartTime: pgtype.Time{
			Microseconds: c.WeekStartTimeMicros,
			Valid:        true,
		},
		RemovedAt:     timestamptz(c.RemovedAt),
		EmojisEnabled: c.EmojisEnabled,
	})
}

func (r *ChatRepository) Get(ctx context.Context, id int64) (chat.Chat, error) {
	u, err := r.queries.GetChatByID(ctx, id)
	if err != nil {
		return chat.Chat{}, err
	}

	return mapChat(u), nil
}

func (r *ChatRepository) SetMentionTypes(ctx context.Context, chatID int64, mentionTypes chat.MentionTypes) error {
	return r.queries.SetChatMentionTypes(ctx, db.SetChatMentionTypesParams{
		MentionTypes: int32(mentionTypes),
		ChatID:       chatID,
	})
}

func (r *ChatRepository) SetSkipSummonConfirmation(ctx context.Context, chatID int64, confirmation bool) error {
	return r.queries.SetChatSkipSummonConfirmation(ctx, db.SetChatSkipSummonConfirmationParams{
		SkipCallConfirmation: confirmation,
		ChatID:               chatID,
	})
}

func (r *ChatRepository) GetUserManagedChats(ctx context.Context, userID int64, search string) ([]chat.Chat, error) {
	chats, err := r.queries.GetUserManagedChats(ctx, db.GetUserManagedChatsParams{
		UserID: userID,
		Title:  search,
	})
	if err != nil {
		return nil, err
	}

	return mapList(chats, mapChat), nil
}

func (r *ChatRepository) GetAllChats(ctx context.Context, search string) ([]chat.Chat, error) {
	chats, err := r.queries.GetAllChats(ctx, search)
	if err != nil {
		return nil, err
	}

	return mapList(chats, mapChat), nil
}

func (r *ChatRepository) SetNewbieThreshold(ctx context.Context, chatID int64, threshold int32) error {
	return r.queries.UpdateChatNewbieThreshold(ctx, db.UpdateChatNewbieThresholdParams{
		NewbieThresholdDays: threshold,
		ID:                  chatID,
	})
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

func (r *ChatRepository) SetWelcomeCallMessage(ctx context.Context, chatID int64, message string) error {
	return r.queries.SetChatWelcomeCallMessage(ctx, db.SetChatWelcomeCallMessageParams{
		WelcomeCallMessage: pgtype.Text{
			String: message,
			Valid:  message != "",
		},
		ChatID: chatID,
	})
}

func (r *ChatRepository) SetCallOnJoin(ctx context.Context, chatID int64, isEnabled bool) error {
	return r.queries.UpdateChatCallOnJoin(ctx, db.UpdateChatCallOnJoinParams{
		CallOnJoin: isEnabled,
		ChatID:     chatID,
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

func (r *ChatRepository) SetTitle(ctx context.Context, chatID int64, title string) error {
	return r.queries.UpdateChatTitle(ctx, db.UpdateChatTitleParams{
		Title: title,
		ID:    chatID,
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

func (r *ChatRepository) SetWeekStartDay(ctx context.Context, chatID int64, day int) error {
	return r.queries.UpdateChatWeekStartDay(ctx, db.UpdateChatWeekStartDayParams{
		ChatID:       chatID,
		WeekStartDay: int16(day),
	})
}

func (r *ChatRepository) SetWeekStartTime(ctx context.Context, chatID int64, timeMicroseconds int64) error {
	return r.queries.UpdateChatWeekStartTime(ctx, db.UpdateChatWeekStartTimeParams{
		ChatID:        chatID,
		WeekStartTime: pgtype.Time{Microseconds: timeMicroseconds, Valid: true},
	})
}

func (r *ChatRepository) SetEmojisEnabled(ctx context.Context, chatID int64, enabled bool) error {
	return r.queries.SetChatEmojisEnabled(ctx, db.SetChatEmojisEnabledParams{
		EmojisEnabled: enabled,
		ID:            chatID,
	})
}
func (r *ChatRepository) ListWithoutTitle(ctx context.Context) ([]chat.Chat, error) {
	chats, err := r.queries.GetChatsWithoutTitle(ctx)
	if err != nil {
		return nil, err
	}
	return mapList(chats, mapChat), nil
}

func (r *ChatRepository) SetPolygamyEnabled(ctx context.Context, chatID int64, enabled bool) error {
	return r.queries.SetAllowPolygamy(ctx, db.SetAllowPolygamyParams{
		AllowPolygamy: enabled,
		ID:            chatID,
	})
}
