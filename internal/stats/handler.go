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

func (h *Handler) Register(registry *command.Registry) {
	addNorm := &command.ActionDef{
		Key:         "add_norm",
		Trigger:     command.TriggerCommand,
		Aliases:     []string{"+норма", "добавить норму"},
		MinStatus:   chatmember.StatusSeniorAdmin,
		Category:    command.CategoryStats,
		Description: i18n.Cmd.AddNorm.Desc,
		Examples:    []i18n.MessageID{i18n.Cmd.AddNorm.ExampleSimple, i18n.Cmd.AddNorm.ExampleNamed, i18n.Cmd.AddNorm.ExampleUsers},
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Args: []predicate.Arg{
			{Type: predicate.ArgTypeNumber, Count: 1, Optional: false},
			{Type: predicate.ArgTypeUser, Count: predicate.ArgCountVariadic, Optional: true},
			{Type: predicate.ArgTypeText, Count: 1, Optional: true},
		},
	}

	registry.Add(addNorm)

	h.bot.OnMessage(h.AddNorm,
		predicate.Command(addNorm.Key, addNorm.Aliases...),
		predicate.WithArgs(h.chatMemberRepository, addNorm.Args...),
	)
}
