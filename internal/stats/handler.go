package stats

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

const CategoryStats command.Category = "stats"

type Handler struct {
	bot         *botapi.Bot
	permissions *predicate.PermissionChecker
	rules       *predicate.RuleChecker

	service *Service
}

func NewHandler(
	b *botapi.Bot,
	p *predicate.PermissionChecker,
	c *predicate.RuleChecker,
	s *Service,
) *Handler {
	return &Handler{
		bot:         b,
		rules:       c,
		permissions: p,
		service:     s,
	}
}

func (h *Handler) Register(registry *command.Registry) {
	registry.AddCategory(CategoryStats)

	chatDef := action.NewCommand(
		"stats",
		i18n.Cmd.Stats.Desc,
		CategoryStats,
		permission.StatusMember,
		option.WithAliases("отчет", "отчёт"),
		option.WithExamples(i18n.Cmd.Stats.ExampleDuration, i18n.Cmd.Stats.ExampleDate),
		option.WithRules(rule.DateTimeOrDuration().Optional()),
	)

	profileDef := action.NewCommand(
		"you",
		i18n.Cmd.Profile.Desc,
		CategoryStats,
		permission.StatusMember,
		option.WithAliases("кто ты", "профиль"),
		option.WithExamples(i18n.Cmd.Profile.ExampleDuration, i18n.Cmd.Profile.ExampleDate),
		option.WithRules(
			rule.User(),
			rule.DateTimeOrDuration().Optional(),
		),
	)

	profileMeDef := action.NewCommand(
		"me",
		i18n.Cmd.Profile.Desc,
		CategoryStats,
		permission.StatusMember,
		option.WithAliases("кто я", "профиль"),
		option.WithExamples(i18n.Cmd.Profile.ExampleDuration, i18n.Cmd.Profile.ExampleDate),
		option.WithRules(rule.DateTimeOrDuration().Optional()),
	)

	registry.Add(chatDef)
	registry.Add(profileDef)
	registry.Add(profileMeDef)

	h.bot.OnMessage(h.Chat,
		predicate.Command(chatDef.Key, chatDef.Aliases...),
		h.rules.With(chatDef.Rules...),
		h.permissions.Require(chatDef.Key, chatDef.MinStatus),
	)

	h.bot.OnMessage(h.Profile,
		predicate.Command(profileDef.Key, profileDef.Aliases...),
		h.rules.With(profileDef.Rules...),
		h.permissions.Require(profileDef.Key, profileDef.MinStatus),
	)

	h.bot.OnMessage(h.Profile,
		predicate.Command(profileMeDef.Key, profileMeDef.Aliases...),
		h.rules.With(profileMeDef.Rules...),
		h.permissions.Require(profileMeDef.Key, profileMeDef.MinStatus),
	)
}
