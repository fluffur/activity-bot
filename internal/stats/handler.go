package stats

import (
	"activity-bot/internal/action"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/rule"
)

const CategoryStats command.Category = "stats"

type Handler struct {
	service *Service
}

func NewHandler(
	s *Service,
) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) Actions() []*command.Action {
	return []*command.Action{
		action.NewCommand(
			"stats",
			h.Chat,
			i18n.Cmd.Stats.Desc,
			CategoryStats,
			option.WithAliases("отчет", "отчёт"),
			option.WithExamples(
				i18n.Cmd.Stats.ExampleDuration,
				i18n.Cmd.Stats.ExampleDate,
			),
			option.WithRules(
				rule.DateTimeOrDuration().Optional(),
			),
		),
		action.NewCommand(
			"top",
			h.Top,
			i18n.Cmd.Top.Desc,
			CategoryStats,
			option.WithAliases("топ", "стата"),
			option.WithRules(
				rule.DateTimeOrDuration().Optional(),
			),
		),

		action.NewCommand(
			"you",
			h.Profile,
			i18n.Cmd.Profile.Desc,
			CategoryStats,
			option.WithAliases("кто ты", "профиль"),
			option.WithExamples(
				i18n.Cmd.Profile.ExampleDuration,
				i18n.Cmd.Profile.ExampleDate,
			),
			option.WithRules(
				rule.User(),
				rule.DateTimeOrDuration().Optional(),
			),
		),

		action.NewCommand(
			"me",
			h.Profile,
			i18n.Cmd.Profile.Desc,
			CategoryStats,
			option.WithAliases("кто я", "профиль"),
			option.WithExamples(
				i18n.Cmd.Profile.ExampleDuration,
				i18n.Cmd.Profile.ExampleDate,
			),
			option.WithRules(
				rule.DateTimeOrDuration().Optional(),
			),
		),
	}
}
