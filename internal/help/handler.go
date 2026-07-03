package help

import (
	"activity-bot/internal/action"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/predicate"
	"activity-bot/internal/rule"

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

	helpDef := action.NewCommand(
		"help",
		i18n.Cmd.Help.Desc,
		CategoryHelp,
		permission.StatusMember,
		option.WithAliases("помощь"),
	)

	helpCommandDef := action.NewCommand(
		"help_command",
		i18n.Cmd.HelpCommand.Desc,
		CategoryHelp,
		permission.StatusMember,
		option.WithAliases("помощь"),
		option.WithRules(rule.Rule{Type: rule.RuleText, CountArgs: 1}),
	)

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
