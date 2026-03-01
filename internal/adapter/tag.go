package adapter

import (
	"activity-bot/internal/chat"
	"context"
	"errors"
	"fmt"

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
	if c.TagsEnabled {
		_, err := a.b.Request("setChatMemberTag", map[string]any{
			"chat_id": chatID,
			"user_id": userID,
			"tag":     tag,
		}, nil)
		if err != nil {
			return err
		}
		return nil
	}

	m, err := a.b.GetChatMember(chatID, userID, nil)
	if err != nil {
		return ErrChatMemberNotFound
	}

	if m.GetStatus() == "creator" {
		return ErrChatMemberIsCreator
	}

	mergedMember := m.MergeChatMember()
	if m.GetStatus() == "administrator" {
		if !mergedMember.CanBeEdited {
			return ErrChatMemberCantBeEdited
		}

		if _, err := a.b.SetChatAdministratorCustomTitle(chatID, userID, tag, nil); err != nil {
			return err
		}
		return nil

	}

	if m.GetStatus() == "member" || m.GetStatus() == "restricted" {
		if ok, err := a.b.PromoteChatMember(chatID, userID, &gotgbot.PromoteChatMemberOpts{
			CanManageChat:   true,
			CanPostMessages: true,
			CanEditMessages: true,
		}); err != nil || !ok {
			return err
		}

		if _, err := a.b.SetChatAdministratorCustomTitle(chatID, userID, tag, nil); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("status %s", m.GetStatus())
}
