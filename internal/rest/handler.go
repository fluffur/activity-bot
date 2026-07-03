package rest

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

const CategoryRest command.Category = "rest"

type Handler struct {
	bot         *botapi.Bot
	permissions *predicate.PermissionChecker
	rules       *predicate.RuleChecker

	chatMemberService *chatmember.Service
	service           *Service
}

func NewHandler(
	b *botapi.Bot,
	p *predicate.PermissionChecker,
	rs *predicate.RuleChecker,
	rsrv *Service,
	cms *chatmember.Service,
) *Handler {
	return &Handler{
		bot:               b,
		permissions:       p,
		rules:             rs,
		service:           rsrv,
		chatMemberService: cms,
	}
}

func (h *Handler) Register(registry *command.Registry) {
	registry.AddCategory(CategoryRest)

	showRestDef := action.NewCommand(
		"rest",
		i18n.Cmd.Rest.Desc,
		CategoryRest,
		permission.StatusMember,
		option.WithAliases("рест", "мой рест"),
		option.WithRules(rule.User().Optional()),
	)

	setRestDef := action.NewCommand(
		"set_rest",
		i18n.Cmd.Rest.Desc,
		CategoryRest,
		permission.StatusModerator,
		option.WithAliases("рест", "+рест", "установить рест"),
		option.WithRules(
			rule.User().Optional(),
			rule.DateTimeOrDuration(),
			rule.Text().Optional(),
		),
	)

	endRestDef := action.NewCommand(
		"end_rest",
		i18n.Cmd.EndRest.Desc,
		CategoryRest,
		permission.StatusModerator,
		option.WithAliases("-рест", "завершить рест"),
		option.WithRules(rule.User().Optional()),
	)

	allRestDef := action.NewCommand(
		"rests",
		i18n.Cmd.Rests.Desc,
		CategoryRest,
		permission.StatusModerator,
		option.WithAliases("ресты", "все ресты"),
		option.WithRules(rule.User().Optional()),
	)

	approveRequestDef := action.NewCallback(
		"approve:",
		CategoryRest,
		permission.StatusModerator,
		setRestDef,
	)

	rejectRequestDef := action.NewCallback(
		"reject:",
		CategoryRest,
		permission.StatusModerator,
		setRestDef,
	)

	removeRequestDef := action.NewCommand(
		"rests_delete",
		i18n.Cmd.Rests.Delete.Desc,
		CategoryRest,
		permission.StatusModerator,
		option.WithAliases("-рест", "-ресты", "удалить рест"),
		option.WithRules(
			rule.User().Optional(),
			rule.Number(),
		),
	)

	registry.Add(showRestDef)
	registry.Add(setRestDef)
	registry.Add(endRestDef)
	registry.Add(allRestDef)
	registry.Add(approveRequestDef)
	registry.Add(rejectRequestDef)
	registry.Add(removeRequestDef)

	h.bot.OnMessage(
		h.ShowRest,
		predicate.Command(showRestDef.Key, showRestDef.Aliases...),
		h.rules.With(showRestDef.Rules...),
		h.permissions.Require(showRestDef.Key, showRestDef.MinStatus),
	)

	h.bot.OnMessage(
		h.SetRest,
		predicate.Command(setRestDef.Key, setRestDef.Aliases...),
		h.rules.With(setRestDef.Rules...),
		h.permissions.Pass(setRestDef.Key, setRestDef.MinStatus),
	)

	h.bot.OnMessage(
		h.EndRest,
		predicate.Command(endRestDef.Key, endRestDef.Aliases...),
		h.rules.With(endRestDef.Rules...),
		h.permissions.Pass(endRestDef.Key, endRestDef.MinStatus),
	)

	h.bot.OnCallbackQuery(
		h.ApproveRestRequest,
		botapi.CallbackPrefix(approveRequestDef.Key),
		h.permissions.Require(approveRequestDef.Key, approveRequestDef.MinStatus),
	)

	h.bot.OnCallbackQuery(
		h.RejectRestRequest,
		botapi.CallbackPrefix(rejectRequestDef.Key),
		h.permissions.Pass(rejectRequestDef.Key, rejectRequestDef.MinStatus),
	)

	h.bot.OnMessage(
		h.AllUserRests,
		predicate.Command(allRestDef.Key, allRestDef.Aliases...),
		h.rules.With(allRestDef.Rules...),
		h.permissions.Require(allRestDef.Key, allRestDef.MinStatus),
	)

	h.bot.OnMessage(
		h.RemoveRestRequest,
		predicate.Command(removeRequestDef.Key, removeRequestDef.Aliases...),
		h.rules.With(removeRequestDef.Rules...),
		h.permissions.Require(removeRequestDef.Key, removeRequestDef.MinStatus),
	)
}
