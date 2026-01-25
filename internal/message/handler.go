package message

import (
	"activity-bot/internal/chat/member"
	"log/slog"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

type Handler struct {
	service       *Service
	memberService *member.Service
}

func NewHandler(service *Service, memberService *member.Service) *Handler {
	return &Handler{service, memberService}
}

func (h *Handler) Message(_ *gotgbot.Bot, ctx *ext.Context) error {
	u := ctx.EffectiveSender.User
	if u.IsBot {
		return nil
	}

	if _, err := h.memberService.EnsureMemberExists(ctx.EffectiveChat.Id, u.Id, u.Username, u.FirstName, u.LastName); err != nil {
		slog.Error("failed to ensure member exists on message", "chat_id", ctx.EffectiveChat.Id, "user_id", u.Id, "error", err)
		return err
	}

	if err := h.service.Save(ctx.EffectiveChat.Id, ctx.EffectiveUser.Id); err != nil {
		slog.Error("failed to save message activity", "chat_id", ctx.EffectiveChat.Id, "user_id", ctx.EffectiveUser.Id, "error", err)
		return err
	}
	return nil
}
