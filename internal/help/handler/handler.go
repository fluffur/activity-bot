package handler

import (
	"activity-bot/internal/cmd"
	"activity-bot/internal/help/view"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

type Handler struct {
	ownerID int64
}

func New(ownerID int64) *Handler {
	return &Handler{ownerID}
}

func (h *Handler) Start(b *gotgbot.Bot, ctx *cmd.Context) error {
	return ctx.Reply(b, view.FormatStartMessage(), nil)
}

func (h *Handler) Help(b *gotgbot.Bot, ctx *cmd.Context) error {
	return ctx.ReplyHTML(b, view.FormatHelpText(h.ownerID))
}
