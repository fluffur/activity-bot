package norm

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

const CategoryNorm command.Category = "norm"

type Handler struct {
	bot         *botapi.Bot
	translator  *i18n.Translator
	rules       *predicate.RuleChecker
	permissions *predicate.PermissionChecker
	repository  Repository
}

func NewHandler(b *botapi.Bot, t *i18n.Translator, p *predicate.PermissionChecker, c *predicate.RuleChecker, nr Repository) *Handler {
	return &Handler{b, t, c, p, nr}
}

func (h *Handler) Register(registry *command.Registry) {
	registry.AddCategory(CategoryNorm)

	addNormDef := &command.ActionDef{
		Key:         "add_norm",
		Trigger:     command.TriggerCommand,
		Aliases:     []string{"+норма", "добавить норму"},
		MinStatus:   chatmember.StatusSeniorAdmin,
		Category:    CategoryNorm,
		Description: i18n.Cmd.AddNorm.Desc,
		Examples:    []i18n.MessageID{i18n.Cmd.AddNorm.ExampleSimple, i18n.Cmd.AddNorm.ExampleNamed, i18n.Cmd.AddNorm.ExampleUsers},
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleNumber, Count: 1, Optional: false},
			{Type: predicate.RuleUser, Count: predicate.RuleVariadic, Optional: true},
			{Type: predicate.RuleText, Count: 1, Optional: true, TextValidate: isValidNormName},
		},
	}

	listNormsDef := &command.ActionDef{
		Key:         "list_norms",
		Trigger:     command.TriggerCommand,
		Aliases:     []string{"нормы"},
		MinStatus:   chatmember.StatusMember,
		Category:    CategoryNorm,
		Description: i18n.Cmd.ListNorms.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
	}

	showNormDef := &command.ActionDef{
		Key:         "norm",
		Trigger:     command.TriggerCommand,
		Aliases:     []string{"норма"},
		MinStatus:   chatmember.StatusMember,
		Category:    CategoryNorm,
		Description: i18n.Cmd.ShowNorm.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleText, Count: 1, Optional: true, TextValidate: isValidNormName},
		},
	}

	deleteNormDef := &command.ActionDef{
		Key:         "delete_norm",
		Trigger:     command.TriggerCommand,
		Aliases:     []string{"-норма", "удалить норму"},
		MinStatus:   chatmember.StatusSeniorAdmin,
		Category:    CategoryNorm,
		Description: i18n.Cmd.DeleteNorm.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleText, Count: 1, Optional: false, TextValidate: isValidNormName},
		},
	}

	assignNormDef := &command.ActionDef{
		Key:         "assign_norm",
		Trigger:     command.TriggerCommand,
		Aliases:     []string{"назначить норму", "назначить", "привязать норму", "привязать"},
		MinStatus:   chatmember.StatusSeniorAdmin,
		Category:    CategoryNorm,
		Description: i18n.Cmd.AssignNorm.Desc,
		Examples:    []i18n.MessageID{i18n.Cmd.AssignNorm.Example},
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleUser, Count: predicate.RuleVariadic, Optional: false},
			{Type: predicate.RuleText, Count: 1, Optional: false, TextValidate: isValidNormName},
		},
	}

	unassignNormDef := &command.ActionDef{
		Key:         "unassign_norm",
		Trigger:     command.TriggerCommand,
		Aliases:     []string{"снять норму", "снять", "отвязать норму", "отвязать"},
		MinStatus:   chatmember.StatusSeniorAdmin,
		Category:    CategoryNorm,
		Description: i18n.Cmd.UnassignNorm.Desc,
		Examples:    []i18n.MessageID{i18n.Cmd.UnassignNorm.Example},
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleUser, Count: predicate.RuleVariadic, Optional: false},
			{Type: predicate.RuleText, Count: 1, Optional: false, TextValidate: isValidNormName},
		},
	}

	registry.Add(addNormDef)
	registry.Add(listNormsDef)
	registry.Add(showNormDef)
	registry.Add(deleteNormDef)
	registry.Add(assignNormDef)
	registry.Add(unassignNormDef)

	h.bot.OnMessage(h.AddNorm,
		predicate.Command(addNormDef.Key, addNormDef.Aliases...),
		h.rules.With(addNormDef.Rules...),
		h.permissions.Require(addNormDef.Key, addNormDef.MinStatus),
	)

	h.bot.OnMessage(
		h.ListNorms,
		predicate.Command(listNormsDef.Key, listNormsDef.Aliases...),
		predicate.NoArgs(),
		h.permissions.Require(deleteNormDef.Key, deleteNormDef.MinStatus),
	)

	h.bot.OnMessage(
		h.ShowNorm,
		predicate.Command(showNormDef.Key, showNormDef.Aliases...),
		h.rules.With(showNormDef.Rules...),
		h.permissions.Require(deleteNormDef.Key, deleteNormDef.MinStatus),
	)

	h.bot.OnMessage(
		h.DeleteNorm,
		predicate.Command(deleteNormDef.Key, deleteNormDef.Aliases...),
		h.rules.With(deleteNormDef.Rules...),
		h.permissions.Require(deleteNormDef.Key, deleteNormDef.MinStatus),
	)

	h.bot.OnMessage(
		h.AssignNorm,
		predicate.Command(assignNormDef.Key, assignNormDef.Aliases...),
		h.rules.With(assignNormDef.Rules...),
		h.permissions.Require(assignNormDef.Key, assignNormDef.MinStatus),
	)

	h.bot.OnMessage(
		h.UnassignNorm,
		predicate.Command(unassignNormDef.Key, unassignNormDef.Aliases...),
		h.rules.With(unassignNormDef.Rules...),
		h.permissions.Require(unassignNormDef.Key, unassignNormDef.MinStatus),
	)
}
