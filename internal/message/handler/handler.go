package handler

import (
	"activity-bot/internal/member"
	"activity-bot/internal/message"
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

func (h *Handler) Message(_ *gotgbot.Bot, ctx *ext.Context) error {
	u := ctx.EffectiveSender.User
	c := ctx.EffectiveChat
	if u == nil || c == nil || u.IsBot {
		return nil
	}

	if _, err := h.memberService.EnsureMemberExists(c.Id, u.Id, u.Username, u.FirstName, u.LastName, "member"); err != nil {
		slog.Error("failed to ensure member exists on message", "chat_id", c.Id, "user_id", u.Id, "error", err)
		return err
	}

	if err := h.service.Save(c.Id, u.Id); err != nil {
		slog.Error("failed to save message activity", "chat_id", c.Id, "user_id", u.Id, "error", err)
		return err
	}
	return nil
}
