package help

import (
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot         *botapi.Bot
	translator  *i18n.Translator
	permissions *predicate.PermissionChecker
	registry    *command.Registry

	commandsURL       string
	developerUsername string
}

func NewHandler(
	b *botapi.Bot,
	t *i18n.Translator,
	p *predicate.PermissionChecker,
	r *command.Registry,
	commandsURL,
	developerUsername string,
) *Handler {
	return &Handler{
		bot:               b,
		translator:        t,
		permissions:       p,
		registry:          r,
		commandsURL:       commandsURL,
		developerUsername: developerUsername,
	}
}
func (h *Handler) Register(registry *command.Registry) {
	helpDef := &command.ActionDef{
		Key:         "help",
		Aliases:     []string{"help", "помощь"},
		Trigger:     command.TriggerCommand,
		Category:    command.CategoryHelp,
		Description: i18n.Cmd.Help.Desc,
		ShowInHelp:  true,
	}

	registry.Add(helpDef)

	h.bot.OnMessage(h.Help,
		predicate.Command(helpDef.Key, helpDef.Aliases...),
		predicate.NoArgs(),
	)

	h.bot.OnCommand("start", "Start bot", h.Help,
		predicate.Private(),
		predicate.NoArgs(),
	)
}
