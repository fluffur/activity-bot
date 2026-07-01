package stats

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

const CategoryStats command.Category = "stats"

type Handler struct {
	bot         *botapi.Bot
	translator  *i18n.Translator
	permissions *predicate.PermissionChecker
	rules       *predicate.RuleChecker

	service   *Service
	presenter *Presenter
}

func NewHandler(
	b *botapi.Bot,
	t *i18n.Translator,
	p *predicate.PermissionChecker,
	c *predicate.RuleChecker,
	s *Service,
	pr *Presenter,
) *Handler {
	return &Handler{
		bot:         b,
		translator:  t,
		rules:       c,
		permissions: p,
		service:     s,
		presenter:   pr,
	}
}

func (h *Handler) Register(registry *command.Registry) {
	registry.AddCategory(CategoryStats)

	chatDef := &command.ActionDef{
		Key:         "stats",
		Aliases:     []string{"отчет", "отчёт"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusMember,
		Category:    CategoryStats,
		Description: i18n.Cmd.Stats.Desc,
		Examples:    []i18n.MessageID{i18n.Cmd.Stats.ExampleDuration, i18n.Cmd.Stats.ExampleDate},
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleDuration, Optional: true, Count: 1},
			{Type: predicate.RuleDateTime, Optional: true, Count: 2},
		},
	}

	profileDef := &command.ActionDef{
		Key:         "you",
		Aliases:     []string{"кто ты", "профиль"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusMember,
		Category:    CategoryStats,
		Description: i18n.Cmd.Profile.Desc,
		Examples:    []i18n.MessageID{i18n.Cmd.Profile.ExampleDuration, i18n.Cmd.Profile.ExampleDate},
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleUser, Optional: false, Count: 1},
			{Type: predicate.RuleDuration, Optional: true, Count: 1},
			{Type: predicate.RuleDateTime, Optional: true, Count: 2},
		},
	}

	profileMeDef := &command.ActionDef{
		Key:         "me",
		Aliases:     []string{"кто я", "профиль"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusMember,
		Category:    CategoryStats,
		Description: i18n.Cmd.Profile.Desc,
		Examples:    []i18n.MessageID{i18n.Cmd.Profile.ExampleDuration, i18n.Cmd.Profile.ExampleDate},
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleDuration, Optional: true, Count: 1},
			{Type: predicate.RuleDateTime, Optional: true, Count: 2},
		},
	}

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
