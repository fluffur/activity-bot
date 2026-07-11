package rest

import (
	"activity-bot/internal/action"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/rule"
)

const CategoryRest command.Category = "rest"

const (
	callbackRestApprove = "rest:approve:"
	callbackRestReject  = "rest:reject:"
)

type Handler struct {
	chatMemberService *chatmember.Service
	service           *Service
}

func NewHandler(
	rsrv *Service,
	cms *chatmember.Service,
) *Handler {
	return &Handler{
		service:           rsrv,
		chatMemberService: cms,
	}
}
func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"rest",
			h.ShowRest,
			i18n.Cmd.Rest.Desc,
			CategoryRest,
			option.WithAliases("рест", "мой рест"),
			option.WithRules(
				rule.User().Optional(),
			),
		),

		action.NewCommand(
			"set_rest",
			h.SetRest,
			i18n.Cmd.Rest.Desc,
			CategoryRest,
			option.WithAliases("рест", "+рест", "установить рест"),
			option.WithPermission(permission.StatusModerator),
			option.IgnorePermissionCheck(),
			option.WithRules(
				rule.User().Optional(),
				rule.DateTimeOrDuration(),
				rule.Text().Optional(),
			),
		),

		action.NewCommand(
			"end_rest",
			h.EndRest,
			i18n.Cmd.EndRest.Desc,
			CategoryRest,
			option.WithAliases("-рест", "завершить рест"),
			option.WithPermission(permission.StatusModerator),
			option.IgnorePermissionCheck(),
			option.WithRules(
				rule.User().Optional(),
			),
		),

		action.NewCommand(
			"rests",
			h.AllUserRests,
			i18n.Cmd.Rests.Desc,
			CategoryRest,
			option.WithAliases("ресты", "все ресты"),
			option.WithRules(
				rule.User().Optional(),
			),
		),

		action.NewCallbackPrefix(
			"approve_rest",
			callbackRestApprove,
			h.ApproveRestRequest,
			CategoryRest,
			option.WithPermission(permission.StatusModerator),
		),

		action.NewCallbackPrefix(
			"reject_rest",
			callbackRestReject,
			h.RejectRestRequest,
			CategoryRest,
			option.WithPermission(permission.StatusModerator),
		),

		action.NewCommand(
			"rests_delete",
			h.RemoveRestRequest,
			i18n.Cmd.Rests.Delete.Desc,
			CategoryRest,
			option.WithPermission(permission.StatusModerator),
			option.WithAliases("-рест", "-ресты", "удалить рест"),
			option.WithRules(
				rule.User().Optional(),
				rule.Number(),
			),
		),
	}
}
