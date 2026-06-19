package help

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot        *botapi.Bot
	translator *i18n.Service

	commandsURL       string
	developerUsername string
}

func NewHandler(bot *botapi.Bot, translator *i18n.Service, commandsURL, developerUsername string) *Handler {
	return &Handler{bot, translator, commandsURL, developerUsername}
}

func (h *Handler) Register() {
	h.bot.OnMessage(h.Help, predicate.Command("help", "помощь"))
}
