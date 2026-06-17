package repository

import (
	"activity-bot/internal/chatmember"
	db "activity-bot/internal/db/sqlc"
	"context"
)

type ChatMemberRepository struct {
	queries *db.Queries
}

func NewChatMemberRepository(queries *db.Queries) chatmember.Repository {
	return &ChatMemberRepository{queries: queries}
}

func (r *ChatMemberRepository) Create(ctx context.Context, m chatmember.ChatMember) error {
	return r.queries.CreateChatMember(ctx, db.CreateChatMemberParams{
		ChatID:     m.Chat.ID,
		UserID:     m.User.ID,
		Tag:        text(m.Tag),
		Status:     int16(m.Status),
		RestUntil:  timestamptz(m.RestUntil),
		LeftAt:     timestamptz(m.LeftAt),
		RestReason: text(m.RestReason),
		EmojiJson:  m.Emojis,
	})

}

func (r *ChatMemberRepository) Get(ctx context.Context, chatID, userID int64) (chatmember.ChatMember, error) {
	m, err := r.queries.GetChatMember(ctx, db.GetChatMemberParams{
		ChatID: chatID,
		UserID: userID,
		IsBot:  false,
	})
	if err != nil {
		return chatmember.ChatMember{}, err
	}

	return mapChatMemberFull(m.ChatMember, m.Chat, m.User), nil
}
