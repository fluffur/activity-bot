package moderation

import (
	"activity-bot/internal/action"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/rule"
)

const CategoryModeration command.Category = "moderation"

type Handler struct {
	service           *Service
	chatMemberService *chatmember.Service
}

func NewHandler(
	service *Service,
	cms *chatmember.Service,
) *Handler {
	return &Handler{
		service:           service,
		chatMemberService: cms,
	}
}

func (h *Handler) Actions() []*command.ActionDef {
	return []*command.ActionDef{
		action.NewCommand(
			"set_role",
			h.SetRole,
			i18n.Cmd.Moderation.SetRole.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusModerator),
			option.WithAliases("+роль", "роль"),
			option.WithRules(
				rule.User().Optional(),
				rule.Text().Validate(isValidRoleString),
			),
		),

		action.NewCommand(
			"set_role_admin",
			h.SetRoleAdmin,
			i18n.Cmd.Moderation.SetRoleAdmin.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusCoOwner),
			option.WithAliases("+адмроль", "адмроль"),
			option.WithRules(
				rule.User().Optional(),
				rule.Text().Validate(isValidRoleString),
			),
		),

		action.NewCommand(
			"ban",
			h.Ban,
			i18n.Cmd.Moderation.Ban.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusSeniorAdmin),
			option.WithAliases("бан", "кик"),
			option.WithRules(
				rule.User(),
				rule.DateTimeOrDuration().Optional(),
				rule.Text().Optional(),
			),
		),

		action.NewCommand(
			"mute",
			h.Mute,
			i18n.Cmd.Moderation.Mute.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusAdmin),
			option.WithAliases("мут", "молчать"),
			option.WithRules(
				rule.User(),
				rule.DateTimeOrDuration().Optional(),
				rule.Text().Optional(),
			),
		),

		action.NewCommand(
			"kick",
			h.Kick,
			i18n.Cmd.Moderation.Kick.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusSeniorAdmin),
			option.WithAliases("кик"),
			option.WithRules(
				rule.User(),
				rule.DateTimeOrDuration().Optional(),
				rule.Text().Optional(),
			),
		),

		action.NewCommand(
			"warn",
			h.Warn,
			i18n.Cmd.Moderation.Warn.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusSeniorAdmin),
			option.WithAliases("варн", "пред"),
			option.WithRules(
				rule.User(),
				rule.DateTimeOrDuration().Optional(),
				rule.Text().Optional(),
			),
		),
	}
}
