package norm

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

const CategoryNorm command.Category = "norm"

type Handler struct {
	bot         *botapi.Bot
	rules       *predicate.RuleChecker
	permissions *predicate.PermissionChecker
	repository  Repository
}

func NewHandler(b *botapi.Bot, p *predicate.PermissionChecker, r *predicate.RuleChecker, nr Repository) *Handler {
	return &Handler{
		bot:         b,
		rules:       r,
		permissions: p,
		repository:  nr,
	}
}

func (h *Handler) Register(registry *command.Registry) {
	registry.AddCategory(CategoryNorm)

	addNormDef := action.NewCommand(
		"add_norm",
		i18n.Cmd.AddNorm.Desc,
		CategoryNorm,
		permission.StatusSeniorAdmin,
		option.WithAliases("+норма", "добавить норму", "норма"),
		option.WithExamples(i18n.Cmd.AddNorm.ExampleSimple, i18n.Cmd.AddNorm.ExampleNamed, i18n.Cmd.AddNorm.ExampleUsers),
		option.WithRules(
			rule.User().Optional().Variadic(),
			rule.Number(),
			rule.Text().Optional().Validate(isValidNormName),
		),
	)

	listNormsDef := action.NewCommand(
		"list_norms",
		i18n.Cmd.ListNorms.Desc,
		CategoryNorm,
		permission.StatusMember,
		option.WithAliases("нормы"),
	)

	showNormDef := action.NewCommand(
		"norm",
		i18n.Cmd.ShowNorm.Desc,
		CategoryNorm,
		permission.StatusMember,
		option.WithAliases("норма какая", "какая норма", "а какая норма", "норма"),
		option.WithRules(rule.Text().Optional().Validate(isValidNormName)),
	)

	deleteNormDef := action.NewCommand(
		"delete_norm",
		i18n.Cmd.DeleteNorm.Desc,
		CategoryNorm,
		permission.StatusSeniorAdmin,
		option.WithAliases("-норма", "удалить норму"),
		option.WithRules(rule.Text().Validate(isValidNormName)),
	)

	assignNormDef := action.NewCommand(
		"assign_norm",
		i18n.Cmd.AssignNorm.Desc,
		CategoryNorm,
		permission.StatusSeniorAdmin,
		option.WithAliases("назначить норму", "назначить", "привязать норму", "привязать"),
		option.WithExamples(i18n.Cmd.AssignNorm.Example),
		option.WithRules(
			rule.User().Variadic(),
			rule.Text().Validate(isValidNormName),
		),
	)

	unassignNormDef := action.NewCommand(
		"unassign_norm",
		i18n.Cmd.UnassignNorm.Desc,
		CategoryNorm,
		permission.StatusSeniorAdmin,
		option.WithAliases("снять норму", "снять", "отвязать норму", "отвязать"),
		option.WithExamples(i18n.Cmd.UnassignNorm.Example),
		option.WithRules(
			rule.User().Variadic(),
			rule.Text().Validate(isValidNormName),
		),
	)

	registry.Add(addNormDef)
	registry.Add(listNormsDef)
	registry.Add(showNormDef)
	registry.Add(deleteNormDef)
	registry.Add(assignNormDef)
	registry.Add(unassignNormDef)

	h.bot.OnMessage(h.AddNorm,
		predicate.Command(addNormDef.Key, addNormDef.Aliases...),
		predicate.SensitiveCommand(),
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
