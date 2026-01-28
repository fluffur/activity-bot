package handler

import (
	"activity-bot/internal/member"
	"activity-bot/internal/message"
	"errors"
	"log/slog"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service       *message.Service
	memberService *member.Service
}

func New(service *message.Service, memberService *member.Service) *Handler {
	return &Handler{service, memberService}
}

func (h *Handler) EnsureMemberCustomTitle(b *gotgbot.Bot, chatID, userID int64) (string, error) {
	m, err := h.memberService.GetMemberTitle(chatID, userID)
	if err != nil && !errors.Is(err, member.ErrInvalidCustomTitle) {
		return "", err
	}

	if m != "" {
		return m, nil
	}

	chatMember, err := b.GetChatMember(chatID, userID, nil)
	if err != nil {
		return "", err
	}

	return chatMember.MergeChatMember().CustomTitle, nil
}

func (h *Handler) Message(b *gotgbot.Bot, ctx *ext.Context) error {
	u := ctx.EffectiveSender.User
	c := ctx.EffectiveChat
	if u == nil || c == nil || u.IsBot {
		return nil
	}

	m, err := h.memberService.EnsureMemberExists(c.Id, u.Id, u.Username, u.FirstName, u.LastName, "member")

	if err != nil {
		slog.Error("failed to ensure member exists on message", "chat_id", c.Id, "user_id", u.Id, "error", err)
		return err
	}

	if m.CustomTitle == "" {
		title, err := h.EnsureMemberCustomTitle(b, c.Id, u.Id)
		if err != nil {
			slog.Error("failed to get member custom title or role", "chat_id", c.Id, "user_id", u.Id, "error", err)
			return err
		}
		if m.CustomTitle != title {
			if err := h.memberService.SetMemberTitle(c.Id, u.Id, &title); err != nil {
				slog.Error("failed to set member custom title", "chat_id", c.Id, "user_id", u.Id, "error", err)
				return err
			}
		}
	}

	if err := h.service.Save(c.Id, u.Id); err != nil {
		slog.Error("failed to save message activity", "chat_id", c.Id, "user_id", u.Id, "error", err)
		return err
	}
	return nil
}
