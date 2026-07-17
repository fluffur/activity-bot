package help

import (
	"activity-bot/internal/action"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/rule"

	"github.com/gotd/botapi"
)

const CategoryHelp command.Category = "help"

const (
	callbackHelpCategories = "help:categories"
	callbackHelpCategory   = "help:category"
	callbackHelpCommand    = "help:command"
)

const commandsPerPage = 8

type Handler struct {
	registry       *command.Registry
	permissionRepo PermissionRepository
}

func NewHandler(
	r *command.Registry,
	permissionRepo PermissionRepository,
) *Handler {
	return &Handler{
		registry:       r,
		permissionRepo: permissionRepo,
	}
}
func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"help",
			h.Help,
			i18n.Cmd.Help.Desc,
			CategoryHelp,
			option.WithScope(command.ScopeAny),
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
			"showhelp",
			h.ShowCommandHelp,
			i18n.Cmd.HelpCommand.Desc,
			CategoryHelp,
			option.WithAliases("помощь", "help"),
			option.WithRules(rule.Text()),
		),
		action.NewCallback(
			"helpcategories",
			callbackHelpCategories,
			h.ShowCategories,
			CategoryHelp,
		),
		action.NewCallbackPrefix(
			"helpcategory",
			callbackHelpCategory,
			h.ShowCategory,
			CategoryHelp,
		),
		action.NewCallbackPrefix(
			"helpcommand",
			callbackHelpCommand,
			h.ShowCommand,
			CategoryHelp,
		),
		action.NewCallback(
			"ignore",
			"ignore",
			func(c *botapi.Context) error {
				return c.AnswerCallback()
			},
			CategoryHelp,
		),
	}
}
