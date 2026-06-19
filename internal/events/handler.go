package events

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/user"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot                  *botapi.Bot
	translator           *i18n.Service
	messageRepository    Repository
	chatRepository       chat.Repository
	userRepository       user.Repository
	chatMemberRepository chatmember.Repository
}

func NewHandler(
	bot *botapi.Bot,
	translator *i18n.Service,
	messageRepository Repository,
	chatRepository chat.Repository,
	userRepository user.Repository,
	chatMemberRepository chatmember.Repository,
) *Handler {
	return &Handler{
		bot,
		translator, messageRepository, chatRepository, userRepository, chatMemberRepository}
}

func (h *Handler) Register() {
	h.bot.Dispatcher().OnChannelParticipant(h.ParticipantUpdate)
	h.bot.OnMessage(h.Message, botapi.Or(
		botapi.ChatTypeIs(botapi.ChatTypeGroup),
		botapi.ChatTypeIs(botapi.ChatTypeSupergroup),
	))
}
