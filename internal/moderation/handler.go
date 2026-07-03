package moderation

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

const CategoryModeration command.Category = "moderation"

type Handler struct {
	bot *botapi.Bot

	rules       *predicate.RuleChecker
	permissions *predicate.PermissionChecker

	chatMemberService *chatmember.Service
}

func NewHandler(
	b *botapi.Bot,
	r *predicate.RuleChecker,
	p *predicate.PermissionChecker,
	cms *chatmember.Service,
) *Handler {
	return &Handler{
		bot:               b,
		rules:             r,
		permissions:       p,
		chatMemberService: cms,
	}
}

func (h *Handler) Register(registry *command.Registry) {
	registry.AddCategory(CategoryModeration)

	setRole := &command.ActionDef{
		Key:         "set_role",
		Aliases:     []string{"+роль", "роль"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusAdmin,
		Category:    CategoryModeration,
		Description: i18n.Cmd.Moderation.SetRole.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleUser, Optional: true, Count: 1},
			{Type: predicate.RuleText, Optional: false, Count: 1},
		},
	}

	setRoleAdmin := &command.ActionDef{
		Key:         "set_role_admin",
		Aliases:     []string{"+адмроль", "адмроль"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusCoOwner,
		Category:    CategoryModeration,
		Description: i18n.Cmd.Moderation.SetRoleAdmin.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleUser, Optional: true, Count: 1},
			{Type: predicate.RuleText, Optional: false, Count: 1},
		},
	}

	registry.Add(setRole)
	registry.Add(setRoleAdmin)

	h.bot.OnMessage(
		h.SetRoleAdmin,
		predicate.Command(setRoleAdmin.Key, setRoleAdmin.Aliases...),
		h.rules.With(setRoleAdmin.Rules...),
		h.permissions.Require(setRoleAdmin.Key, setRoleAdmin.MinStatus),
	)
	h.bot.OnMessage(
		h.SetRole,
		predicate.Command(setRole.Key, setRole.Aliases...),
		h.rules.With(setRole.Rules...),
		h.permissions.Require(setRole.Key, setRole.MinStatus),
	)

}
