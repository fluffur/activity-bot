package events

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/info"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot           *botapi.Bot
	translator    *i18n.Translator
	memberService *chatmember.Service
	roleUpdater   *info.Updater
}

func NewHandler(
	b *botapi.Bot,
	t *i18n.Translator,
	ms *chatmember.Service,
	roleUpdater *info.Updater,
) *Handler {
	return &Handler{
		bot:           b,
		translator:    t,
		memberService: ms,
		roleUpdater:   roleUpdater,
	}
}

func (h *Handler) Attach() {
	h.bot.Dispatcher().OnChannelParticipant(h.ParticipantUpdate)
	h.bot.Dispatcher().OnPendingJoinRequests(h.PendingJoinRequests)
}
