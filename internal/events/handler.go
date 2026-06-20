package events

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/message"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot               *botapi.Bot
	translator        *i18n.Translator
	messageRepository message.Repository
	memberService     *chatmember.Service
}

func NewHandler(
	bot *botapi.Bot, t *i18n.Translator, mr message.Repository, ms *chatmember.Service,
) *Handler {
	return &Handler{bot, t, mr, ms}
}

func (h *Handler) Register() {
	h.bot.Dispatcher().OnChannelParticipant(h.ParticipantUpdate)
	h.bot.OnMessage(h.Message, botapi.Or(
		botapi.ChatTypeIs(botapi.ChatTypeGroup),
		botapi.ChatTypeIs(botapi.ChatTypeSupergroup),
	))
}
