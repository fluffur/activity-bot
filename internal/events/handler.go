package events

import (
	"activity-bot/internal/i18n"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot        *botapi.Bot
	translator *i18n.Service
	repository Repository
}

func NewHandler(bot *botapi.Bot, translator *i18n.Service, repository Repository) *Handler {
	return &Handler{bot, translator, repository}
}

func (h *Handler) Register() {
	h.bot.Dispatcher().OnChannelParticipant(h.ParticipantLeft)

	h.bot.OnMessage(h.Message, botapi.Or(
		botapi.ChatTypeIs(botapi.ChatTypeGroup),
		botapi.ChatTypeIs(botapi.ChatTypeSupergroup),
	))
}
