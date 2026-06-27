package events

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot           *botapi.Bot
	translator    *i18n.Translator
	memberService *chatmember.Service
}

func NewHandler(bot *botapi.Bot, t *i18n.Translator, ms *chatmember.Service) *Handler {
	return &Handler{bot, t, ms}
}

func (h *Handler) Register() {
	h.bot.Dispatcher().OnChannelParticipant(h.ParticipantUpdate)
}
