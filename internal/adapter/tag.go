package adapter

import (
	"activity-bot/internal/chat"
	"context"
	"errors"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type MemberTagAdapter struct {
	b              *gotgbot.Bot
	chatRepository chat.Repository
}

func NewMemberTagAdapter(b *gotgbot.Bot, chatRepository chat.Repository) *MemberTagAdapter {
	return &MemberTagAdapter{b, chatRepository}
}

var (
	ErrChatMemberNotFound     = errors.New("chat member not found")
	ErrChatMemberIsCreator    = errors.New("chat member is creator")
	ErrChatMemberCantBeEdited = errors.New("chat member cant be edited")
	ErrChatMemberIsRestricted = errors.New("chat member is restricted")
)

func (a *MemberTagAdapter) SetMemberTag(ctx context.Context, chatID int64, userID int64, tag string) error {
	c, err := a.chatRepository.GetChat(ctx, chatID)
	if err != nil {
		return err
	}
	if _, err := a.b.SetChatMemberTag(chatID, userID, &gotgbot.SetChatMemberTagOpts{
		Tag: tag,
	}); err != nil {
		return err
	}

	if c.TagsEnabled {
		return nil
	}

	m, err := a.b.GetChatMember(chatID, userID, nil)
	if err != nil {
		return ErrChatMemberNotFound
	}

	if m.GetStatus() == "member" || m.GetStatus() == "restricted" {
		if ok, err := a.b.PromoteChatMember(chatID, userID, &gotgbot.PromoteChatMemberOpts{
			CanManageChat:   true,
			CanPostMessages: true,
			CanEditMessages: true,
		}); err != nil || !ok {
			return err
		}
	}

	return nil
}
