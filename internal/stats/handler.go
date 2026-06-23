package stats

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot                  *botapi.Bot
	translator           *i18n.Translator
	chatMemberRepository chatmember.Repository
}

func NewHandler(b *botapi.Bot, t *i18n.Translator, cmr chatmember.Repository) *Handler {
	return &Handler{b, t, cmr}
}

// +норма [название] <число>
// [@user1 @user2 @user3]
//

func (h *Handler) Register(registry *command.Registry) {
	registry.Add(&command.ActionDef{
		Key:         "add_norm",
		Trigger:     command.TriggerCommand,
		Aliases:     []string{"+норма", "добавить норму"},
		MinStatus:   chatmember.StatusSeniorAdmin,
		Category:    command.CategoryStats,
		Description: i18n.AddNormDescription,
		Examples:    []i18n.MessageID{i18n.AddNormExampleSimple, i18n.AddNormExampleNamed, i18n.AddNormExampleAdmins, i18n.AddNormExampleUsers},
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
	})

	h.bot.OnMessage(h.AddNorm, predicate.Command("add_norm", "+норма", "добавить норму"))
}
