package help

import (
	"activity-bot/internal/action"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/rule"
)

const CategoryHelp command.Category = "help"

const (
	callbackHelpCategories = "help:categories"
	callbackHelpCategory   = "help:category"
	callbackHelpCommand    = "help:command"
)

type Handler struct {
	registry *command.Registry

	commandsURL       string
	developerUsername string
}

func NewHandler(
	r *command.Registry,
	commandsURL,
	developerUsername string,
) *Handler {
	return &Handler{
		registry:          r,
		commandsURL:       commandsURL,
		developerUsername: developerUsername,
	}
}
func (h *Handler) Actions() []*command.ActionDef {
	return []*command.ActionDef{
		action.NewCommand(
			"help",
			h.Help,
			i18n.Cmd.Help.Desc,
			CategoryHelp,
			option.WithAliases("помощь"),
		),
		action.NewCommand(
			"start",
			h.Help,
			i18n.Cmd.Help.Desc,
			CategoryHelp,
			option.WithScope(command.ScopePrivate),
		),
		action.NewCommand(
			"help_command",
			h.ShowCommandHelp,
			i18n.Cmd.HelpCommand.Desc,
			CategoryHelp,
			option.WithAliases("помощь"),
			option.WithRules(rule.Text()),
		),
		action.NewCallback(
			"help_categories",
			callbackHelpCategories,
			h.ShowCategories,
			CategoryHelp,
		),
		action.NewCallbackPrefix(
			"help_category",
			callbackHelpCategory,
			h.ShowCategory,
			CategoryHelp,
		),
		action.NewCallbackPrefix(
			"help_command_callback",
			callbackHelpCommand,
			h.ShowCommand,
			CategoryHelp,
		),
	}
}
