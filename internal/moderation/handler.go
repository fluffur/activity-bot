package moderation

import (
	"activity-bot/internal/action"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/predicate"
	"activity-bot/internal/rule"

	"github.com/gotd/botapi"
)

const CategoryModeration command.Category = "moderation"

type Handler struct {
	bot *botapi.Bot

	rules       *predicate.RuleChecker
	permissions *predicate.PermissionChecker

	service           *Service
	chatMemberService *chatmember.Service
}

func NewHandler(
	b *botapi.Bot,
	r *predicate.RuleChecker,
	p *predicate.PermissionChecker,
	service *Service,
	cms *chatmember.Service,
) *Handler {
	return &Handler{
		bot:               b,
		rules:             r,
		permissions:       p,
		service:           service,
		chatMemberService: cms,
	}
}

func (h *Handler) Register(registry *command.Registry) {
	registry.AddCategory(CategoryModeration)

	setRole := action.NewCommand(
		"set_role",
		i18n.Cmd.Moderation.SetRole.Desc,
		CategoryModeration,
		permission.StatusAdmin,
		option.WithAliases("+роль", "роль"),
		option.WithRules(
			rule.User().Optional(),
			rule.Text().Validate(isValidRoleString),
		),
	)

	setRoleAdmin := action.NewCommand(
		"set_role_admin",
		i18n.Cmd.Moderation.SetRoleAdmin.Desc,
		CategoryModeration,
		permission.StatusCoOwner,
		option.WithAliases("+адмроль", "адмроль"),
		option.WithRules(
			rule.User().Optional(),
			rule.Text().Validate(isValidRoleString),
		),
	)

	ban := action.NewCommand(
		"ban",
		i18n.Cmd.Moderation.Ban.Desc,
		CategoryModeration,
		permission.StatusSeniorAdmin,
		option.WithAliases("бан", "кик"),
		option.WithRules(
			rule.User(),
			rule.DateTimeOrDuration().Optional(),
			rule.Text().Optional(),
		),
	)

	registry.Add(setRole)
	registry.Add(setRoleAdmin)
	registry.Add(ban)

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
	h.bot.OnMessage(
		h.Ban,
		predicate.Command(ban.Key, ban.Aliases...),
		h.rules.With(ban.Rules...),
		h.permissions.Require(ban.Key, ban.MinStatus),
	)

}
