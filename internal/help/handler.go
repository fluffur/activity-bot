package help

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot         *botapi.Bot
	translator  *i18n.Translator
	permissions *predicate.PermissionChecker

	commandsURL       string
	developerUsername string
}

func NewHandler(b *botapi.Bot, t *i18n.Translator, p *predicate.PermissionChecker, commandsURL, developerUsername string) *Handler {
	return &Handler{b, t, p, commandsURL, developerUsername}
}
func (h *Handler) Register() {
	h.bot.OnMessage(h.Help,
		predicate.Command("help", "помощь"),
	)

	h.bot.OnCommand("start", "Start bot", h.Help, predicate.Private())
}
