package norm

import (
	"activity-bot/internal/action"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/rule"
)

const CategoryNorm command.Category = "norm"

type Handler struct {
	repository Repository
}

func NewHandler(
	nr Repository,
) *Handler {
	return &Handler{
		repository: nr,
	}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"add_norm",
			h.AddNorm,
			i18n.Cmd.AddNorm.Desc,
			CategoryNorm,
			option.WithPermission(permission.StatusSeniorAdmin),
			option.WithAliases("+норма", "добавить норму", "норма"),
			option.WithExamples(
				i18n.Cmd.AddNorm.ExampleSimple,
				i18n.Cmd.AddNorm.ExampleNamed,
				i18n.Cmd.AddNorm.ExampleUsers,
			),
			option.WithRules(
				rule.User().Optional().Variadic(),
				rule.Number(),
				rule.Text().Optional().Validate(isValidNormName),
			),
		),

		action.NewCommand(
			"list_norms",
			h.ListNorms,
			i18n.Cmd.ListNorms.Desc,
			CategoryNorm,
			option.WithAliases("нормы"),
		),

		action.NewCommand(
			"norm",
			h.ShowNorm,
			i18n.Cmd.ShowNorm.Desc,
			CategoryNorm,
			option.WithAliases(
				"норма какая",
				"какая норма",
				"а какая норма",
				"норма",
			),
			option.WithRules(
				rule.Text().Optional().Validate(isValidNormName),
			),
		),

		action.NewCommand(
			"delete_norm",
			h.DeleteNorm,
			i18n.Cmd.DeleteNorm.Desc,
			CategoryNorm,
			option.WithPermission(permission.StatusSeniorAdmin),
			option.WithAliases("-норма", "удалить норму"),
			option.WithRules(
				rule.Text().Optional().Validate(isValidNormName),
			),
		),

		action.NewCommand(
			"assign_norm",
			h.AssignNorm,
			i18n.Cmd.AssignNorm.Desc,
			CategoryNorm,
			option.WithPermission(permission.StatusSeniorAdmin),
			option.WithAliases(
				"назначить норму",
				"назначить",
				"привязать норму",
				"привязать",
			),
			option.WithExamples(i18n.Cmd.AssignNorm.Example),
			option.WithRules(
				rule.User().Variadic(),
				rule.Text().Validate(isValidNormName),
			),
		),

		action.NewCommand(
			"unassign_norm",
			h.UnassignNorm,
			i18n.Cmd.UnassignNorm.Desc,
			CategoryNorm,
			option.WithPermission(permission.StatusSeniorAdmin),
			option.WithAliases(
				"снять норму",
				"отвязать норму",
				"отвязать",
			),
			option.WithExamples(i18n.Cmd.UnassignNorm.Example),
			option.WithRules(
				rule.User().Variadic(),
				rule.Text().Validate(isValidNormName),
			),
		),
	}
}
