package events

import (
	"activity-bot/internal/application"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/info"
	"activity-bot/internal/summon"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot           *botapi.Bot
	translator    *i18n.Translator
	memberService *chatmember.Service
	roleUpdater   *info.Updater

	applicationRepository *application.Repository
	summonHandler         *summon.Handler
}

func NewHandler(
	b *botapi.Bot,
	t *i18n.Translator,
	ms *chatmember.Service,
	roleUpdater *info.Updater,
	applicationRepository *application.Repository,
	summonHandler *summon.Handler,
) *Handler {
	return &Handler{
		bot:                   b,
		translator:            t,
		memberService:         ms,
		roleUpdater:           roleUpdater,
		applicationRepository: applicationRepository,
		summonHandler:         summonHandler,
	}
}

func (h *Handler) Attach() {
	h.bot.Dispatcher().OnChannelParticipant(h.ParticipantUpdate)
	h.bot.Dispatcher().OnBotChatInviteRequester(h.BotChatInviteRequester)
}
