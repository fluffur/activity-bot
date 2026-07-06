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

func NewHandler(
	b *botapi.Bot,
	t *i18n.Translator,
	ms *chatmember.Service,
) *Handler {
	return &Handler{
		bot:           b,
		translator:    t,
		memberService: ms,
	}
}

func (h *Handler) Attach() {
	h.bot.Dispatcher().OnChannelParticipant(h.ParticipantUpdate)
}
