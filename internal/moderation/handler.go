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

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
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
			"roles",
			h.ListRoles,
			i18n.Cmd.Moderation.ListRoles.Desc,
			CategoryModeration,
			option.WithAliases("роли"),
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
			"unban",
			h.Unban,
			i18n.Cmd.Moderation.Unban.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusAdmin),
			option.WithAliases("разбан", "размут", "снять мут", "говори"),
			option.WithRules(
				rule.User(),
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
			"promote",
			h.Promote,
			i18n.Cmd.Moderation.Promote.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusCoOwner),
			option.WithAliases("повысить"),
			option.WithRules(rule.User(), rule.Number().Optional()),
		),
		action.NewCommand(
			"demote",
			h.Demote,
			i18n.Cmd.Moderation.Demote.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusCoOwner),
			option.WithAliases("понизить"),
			option.WithRules(rule.User(), rule.Number().Optional()),
		),
		action.NewCommand(
			"set_status",
			h.SetStatus,
			i18n.Cmd.Moderation.SetStatus.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusCoOwner),
			option.AllowDev(),
			option.WithAliases("статус", "админ", "+админ", "права"),
			option.WithRules(rule.User(), rule.Number().Optional()),
			option.WithPredicates(predicate.SensitiveCommand()),
		),

		action.NewCommand(
			"admins",
			h.ListAdmins,
			i18n.Cmd.Moderation.ListAdmins.Desc,
			CategoryModeration,
			option.WithAliases("админы", "кто здесь власть"),
		),
		action.NewCommand(
			"warn",
			h.Warn,
			i18n.Cmd.Moderation.Warn.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusAdmin),
			option.WithAliases("варн", "пред"),
			option.WithRules(
				rule.User(),
				rule.DateTimeOrDuration().Optional(),
				rule.Text().Optional(),
			),
		),
		action.NewCommand(
			"unwarn",
			h.Unwarn,
			i18n.Cmd.Moderation.Unwarn.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusSeniorAdmin),
			option.WithAliases("-пред", "снять пред", "-варн", "снять варн"),
			option.WithRules(
				rule.User(),
			),
		),
		action.NewCommand(
			"clear_warns",
			h.ClearWarns,
			i18n.Cmd.Moderation.ClearWarns.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusAdmin),
			option.WithAliases("очистить преды", "очистить варны"),
			option.WithRules(
				rule.User(),
			),
		),
		action.NewCommand(
			"show_warns",
			h.ShowWarns,
			i18n.Cmd.Moderation.ShowWarns.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusModerator),
			option.WithAliases("показать варны", "варны"),
			option.WithRules(
				rule.User().Optional(),
			),
		),
		action.NewCommand(
			"max_warns",
			h.ShowMaxWarns,
			i18n.Cmd.Moderation.MaxWarns.Desc,
			CategoryModeration,
			option.WithAliases("макс преды", "макс варны"),
		),

		action.NewCommand(
			"set_max_warns",
			h.SetMaxWarns,
			i18n.Cmd.Moderation.SetMaxWarns.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusCoOwner),
			option.WithAliases("max_warns", "макс преды", "макс варны"),
			option.WithRules(
				rule.Number(),
			),
			option.WithPredicates(predicate.SensitiveCommand()),
		),

		action.NewCommand(
			"warnlist",
			h.WarnList,
			i18n.Cmd.Moderation.WarnList.Desc,
			CategoryModeration,
			option.WithPermission(permission.StatusModerator),
			option.WithAliases("варнлист"),
		),
	}
}
