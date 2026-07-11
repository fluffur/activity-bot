package handler

import (
	"activity-bot/internal/action"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/rule"
)

const CategoryPermission command.Category = "permission"

type Handler struct {
	registry *command.Registry
	repo     permission.Repository
}

func NewHandler(registry *command.Registry, repo permission.Repository) *Handler {
	return &Handler{
		registry: registry,
		repo:     repo,
	}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"set_permission",
			h.SetPermission,
			i18n.Cmd.Permission.Set.Desc,
			CategoryPermission,
			option.WithAliases("дк", "+дк"),
			option.WithRules(rule.Number(), rule.Text()),
		),
		action.NewCommand(
			"show_permission",
			h.ShowPermission,
			i18n.Cmd.Permission.Show.Desc,
			CategoryPermission,
			option.WithAliases("дк"),
			option.WithRules(rule.Text()),
		),
	}
}
