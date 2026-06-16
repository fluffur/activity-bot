package events

import (
	"activity-bot/internal/i18n"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot        *botapi.Bot
	translator *i18n.Service
	service    *Service
}

func NewHandler(bot *botapi.Bot, translator *i18n.Service, service *Service) *Handler {
	return &Handler{bot, translator, service}
}

func (h *Handler) Register() {
	h.bot.OnMessage(h.Message, botapi.Or(
		botapi.ChatTypeIs(botapi.ChatTypeGroup),
		botapi.ChatTypeIs(botapi.ChatTypeSupergroup),
	))
	h.bot.Dispatcher().OnChannelParticipant(h.ParticipantLeft)
}
