package summon

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot               *botapi.Bot
	translator        *i18n.Translator
	permissions       *predicate.PermissionChecker
	chatMemberService *chatmember.Service
}

func NewHandler(b *botapi.Bot, t *i18n.Translator, p *predicate.PermissionChecker, cms *chatmember.Service) *Handler {
	return &Handler{b, t, p, cms}
}

func (h *Handler) Register() {
	h.bot.OnMessage(h.Summon,
		predicate.Command("summon", "call", "калл", "колл", "каллалл"),
		h.permissions.Require("summon", chatmember.StatusAdmin),
	)
}
