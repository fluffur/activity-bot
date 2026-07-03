package help

import (
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

const CategoryHelp command.Category = "help"

type Handler struct {
	bot         *botapi.Bot
	rules       *predicate.RuleChecker
	permissions *predicate.PermissionChecker
	registry    *command.Registry

	commandsURL       string
	developerUsername string
}

func NewHandler(
	b *botapi.Bot,
	rules *predicate.RuleChecker,
	p *predicate.PermissionChecker,
	r *command.Registry,
	commandsURL,
	developerUsername string,
) *Handler {
	return &Handler{
		bot:               b,
		permissions:       p,
		rules:             rules,
		registry:          r,
		commandsURL:       commandsURL,
		developerUsername: developerUsername,
	}
}
func (h *Handler) Register(registry *command.Registry) {
	registry.AddCategory(CategoryHelp)

	helpDef := &command.ActionDef{
		Key:         "help",
		Aliases:     []string{"help", "помощь"},
		Trigger:     command.TriggerCommand,
		Category:    CategoryHelp,
		Description: i18n.Cmd.Help.Desc,
		ShowInHelp:  true,
	}

	helpCommandDef := &command.ActionDef{
		Key:      "help_command",
		Aliases:  []string{"help", "помощь"},
		Trigger:  command.TriggerCommand,
		Category: CategoryHelp,
		Rules: []predicate.Rule{
			{Type: predicate.RuleText, Count: 1},
		},
		Description: i18n.Cmd.HelpCommand.Desc,
		ShowInHelp:  true,
	}

	registry.Add(helpDef)
	registry.Add(helpCommandDef)

	h.bot.OnMessage(
		h.ShowCommandHelp,
		predicate.Command(helpCommandDef.Key, helpCommandDef.Aliases...),
		h.rules.With(helpCommandDef.Rules...),
	)

	h.bot.OnMessage(
		h.Help,
		predicate.Command(helpDef.Key, helpDef.Aliases...),
		predicate.NoArgs(),
	)

	h.bot.OnCommand(
		"start",
		"Start bot",
		h.Help,
		predicate.Private(),
		predicate.NoArgs(),
	)

	h.bot.OnCallbackQuery(
		h.ShowCategories,
		botapi.CallbackData(callbackHelpCategories),
	)

	h.bot.OnCallbackQuery(
		h.ShowCategory,
		botapi.CallbackPrefix(callbackHelpCategory+":"),
	)

	h.bot.OnCallbackQuery(
		h.ShowCommand,
		botapi.CallbackPrefix(callbackHelpCommand+":"),
	)

}
