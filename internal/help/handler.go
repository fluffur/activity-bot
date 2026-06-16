package help

import (
	"activity-bot/internal/i18n"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot        *botapi.Bot
	translator *i18n.Service
}

func NewHandler(bot *botapi.Bot, translator *i18n.Service) *Handler {
	return &Handler{bot, translator}
}

func (h *Handler) Register() {
	h.bot.OnCommand("help", "", h.Help)
}
