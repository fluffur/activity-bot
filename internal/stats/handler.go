package stats

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

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
	chatDef := &command.ActionDef{
		Key:         "stats",
		Aliases:     []string{"отчет", "отчёт"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusMember,
		Category:    command.CategoryStats,
		Description: "",
		Examples:    nil,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleDuration, Optional: true, Count: 1},
			{Type: predicate.RuleDateTime, Optional: true, Count: 2},
		},
	}
	registry.Add(chatDef)

	h.bot.OnMessage(h.Chat,
		predicate.Command(chatDef.Key, chatDef.Aliases...),
		h.rules.With(chatDef.Rules...),
		h.permissions.Require(chatDef.Key, chatDef.MinStatus),
	)
}
