package handler

import (
	"activity-bot/internal/call"
	"activity-bot/internal/call/view"
	"activity-bot/internal/chat"
	"activity-bot/internal/cmd"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	service     *call.Service
	chatService *chat.Service
}

func New(service *call.Service, chatService *chat.Service) *Handler {
	return &Handler{service, chatService}
}

func (h *Handler) Call(b *gotgbot.Bot, ctx *cmd.Context) error {
	return h.service.Call(ctx.StdContext(), b, ctx.Context, ctx.FirstArgument())
}

func (h *Handler) SetWelcomeCallMessage(b *gotgbot.Bot, ctx *cmd.Context) error {
	message := ctx.FirstArgument()
	if err := h.service.SetWelcomeCallMessage(ctx.StdContext(), ctx.EffectiveChat.Id, message); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatWelcomeCallMessageSet(), nil)
}

func (h *Handler) EnableCallOnJoin(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.EnableCallOnJoin(ctx.StdContext(), ctx.EffectiveChat.Id); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatCallOnJoinEnabled(), nil)
}

func (h *Handler) DisableCallOnJoin(b *gotgbot.Bot, ctx *cmd.Context) error {
	if err := h.service.DisableCallOnJoin(ctx.StdContext(), ctx.EffectiveChat.Id); err != nil {
		return err
	}

	return ctx.Reply(b, view.FormatCallOnJoinDisabled(), nil)
}

func (h *Handler) ShowWelcomeCallMessage(b *gotgbot.Bot, ctx *cmd.Context) error {
	c, err := h.chatService.GetChat(ctx.StdContext(), ctx.EffectiveChat.Id)
	if err != nil {
		return err
	}
	return ctx.Reply(b, view.FormatWelcomeCallMessage(c.WelcomeCallMessage), nil)
}
