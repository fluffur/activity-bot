package manage

import (
	"activity-bot/internal/action"
	"activity-bot/internal/chat"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/pmsession"
	"activity-bot/internal/rule"
)

type Handler struct {
	chatService       *chat.Service
	sessionRepository pmsession.Repository
}

func NewHandler(
	cs *chat.Service,
	sr pmsession.Repository,
) *Handler {
	return &Handler{
		chatService:       cs,
		sessionRepository: sr,
	}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"manage",
			h.Manage,
			i18n.Cmd.Manage.Desc,
			"",
			option.WithAliases("управление"),
			option.WithRules(rule.Text().Optional()),
			option.WithScope(command.ScopePrivate),
			option.Hidden(),
		),

		action.NewCallbackPrefix(
			"manage_callback",
			"manage:",
			h.CallbackManage,
			"",
		),
	}
}
