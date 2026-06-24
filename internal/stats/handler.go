package stats

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/norm"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot            *botapi.Bot
	translator     *i18n.Translator
	args           *predicate.ArgChecker
	permissions    *predicate.PermissionChecker
	normRepository norm.Repository
}

func NewHandler(b *botapi.Bot, t *i18n.Translator, p *predicate.PermissionChecker, c *predicate.ArgChecker, nr norm.Repository) *Handler {
	return &Handler{b, t, c, p, nr}
}

func (h *Handler) Register(registry *command.Registry) {
	addNormDef := &command.ActionDef{
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

	listNormsDef := &command.ActionDef{
		Key:         "list_norms",
		Trigger:     command.TriggerCommand,
		Aliases:     []string{"нормы"},
		MinStatus:   chatmember.StatusMember,
		Category:    command.CategoryStats,
		Description: i18n.Cmd.ListNorms.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
	}

	showNormDef := &command.ActionDef{
		Key:         "norm",
		Trigger:     command.TriggerCommand,
		Aliases:     []string{"норма"},
		MinStatus:   chatmember.StatusMember,
		Category:    command.CategoryStats,
		Description: i18n.Cmd.ShowNorm.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Args: []predicate.Arg{
			{Type: predicate.ArgTypeText, Count: 1, Optional: true},
		},
	}

	deleteNormDef := &command.ActionDef{
		Key:         "delete_norm",
		Trigger:     command.TriggerCommand,
		Aliases:     []string{"-норма", "удалить норму"},
		MinStatus:   chatmember.StatusSeniorAdmin,
		Category:    command.CategoryStats,
		Description: i18n.Cmd.DeleteNorm.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Args: []predicate.Arg{
			{Type: predicate.ArgTypeText, Count: 1, Optional: false},
		},
	}

	registry.Add(addNormDef)
	registry.Add(listNormsDef)
	registry.Add(showNormDef)
	registry.Add(deleteNormDef)

	h.bot.OnMessage(h.AddNorm,
		predicate.Command(addNormDef.Key, addNormDef.Aliases...),
		h.args.WithArgs(addNormDef.Args...),
		h.permissions.Require(addNormDef.Key, addNormDef.MinStatus),
	)

	h.bot.OnMessage(
		h.ListNorms,
		predicate.Command(listNormsDef.Key, listNormsDef.Aliases...),
	)

	h.bot.OnMessage(
		h.ShowNorm,
		predicate.Command(showNormDef.Key, showNormDef.Aliases...),
		h.args.WithArgs(showNormDef.Args...),
	)

	h.bot.OnMessage(
		h.DeleteNorm,
		predicate.Command(deleteNormDef.Key, deleteNormDef.Aliases...),
		h.args.WithArgs(deleteNormDef.Args...),
		h.permissions.Require(deleteNormDef.Key, deleteNormDef.MinStatus),
	)
}
