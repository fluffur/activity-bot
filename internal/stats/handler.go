package stats

import (
	"activity-bot/internal/action"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/norm"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/rule"
	"activity-bot/internal/summon"

	fsm "github.com/fluffur/botapi-fsm"
)

const CategoryStats command.Category = "stats"

type Handler struct {
	service  *Service
	normRepo norm.Repository
	summonH  *summon.Handler
	statsFSM *fsm.Machine[State, StateData]
}

func NewHandler(
	s *Service,
	normRepo norm.Repository,
	summonH *summon.Handler,
	statsFSM *fsm.Machine[State, StateData],
) *Handler {
	return &Handler{
		service:  s,
		normRepo: normRepo,
		summonH:  summonH,
		statsFSM: statsFSM,
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
				rule.Text().Optional().Validate(isAllTime),
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
				rule.Text().Optional().Validate(isAllTime),
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
		),
		action.NewCommand(
			"inactive",
			h.ListInactive,
			i18n.Cmd.Inactive.Desc,
			CategoryStats,
			option.WithAliases("неактив", "инактив", "кто неактив"),
			option.WithRules(rule.Duration().Optional()),
		),
		action.NewCallbackPrefix(
			"summon_no_norm",
			"summon_no_norm",
			h.AskForNormName,
			CategoryStats,
			option.WithPermission(permission.StatusAdmin),
		),
		action.NewCommand(
			"cancel",
			h.Cancel,
			"",
			CategoryStats,
			option.WithAliases("отмена", "отменить"),
			option.WithPredicates(h.statsFSM.InState(StateAwaitNorm, StateAwaitSummonText, StateAwaitInactiveSummonText)),
			option.Hidden(),
		),
		action.NewCallbackPrefix(
			"summonnormselect",
			"summon:norm:",
			h.ProcessNormNameCallback,
			CategoryStats,
			option.WithPredicates(h.statsFSM.State(StateAwaitNorm)),
		),
		action.NewMessage(
			h.ProcessSummonText,
			CategoryStats,
			option.WithPredicates(h.statsFSM.State(StateAwaitSummonText)),
		),
		action.NewCallbackPrefix(
			"summon_inactive",
			"summon_inactive:",
			h.AskInactiveSummonText,
			CategoryStats,
			option.WithPermission(permission.StatusAdmin),
		),
		action.NewMessage(
			h.ProcessInactiveSummonText,
			CategoryStats,
			option.WithPredicates(h.statsFSM.State(StateAwaitInactiveSummonText)),
		),
	}
}
