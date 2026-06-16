package events

import "github.com/gotd/botapi"

type Handler struct {
	bot     *botapi.Bot
	service *Service
}

func NewHandler(bot *botapi.Bot, service *Service) *Handler {
	return &Handler{bot, service}
}

func (h *Handler) Register() {
	h.bot.OnMessage(h.Message, botapi.Or(
		botapi.ChatTypeIs(botapi.ChatTypeGroup),
		botapi.ChatTypeIs(botapi.ChatTypeSupergroup),
	))
	h.bot.Dispatcher().OnChannelParticipant(h.ParticipantLeft)
}
