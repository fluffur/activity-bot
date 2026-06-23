package summon

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
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

func (h *Handler) Register(registry *command.Registry) {
	registry.Add(&command.ActionDef{
		Key:         "summon",
		Aliases:     []string{"call", "калл", "колл", "каллалл"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusAdmin,
		Category:    command.CategorySummon,
		Description: i18n.Cmd.Summon.Desc,
		Examples:    []i18n.MessageID{},
		ShowInHelp:  true,
	})

	h.bot.OnMessage(h.Summon,
		predicate.Command("summon", "call", "калл", "колл", "каллалл"),
		h.permissions.Require("summon", chatmember.StatusAdmin),
	)
}
