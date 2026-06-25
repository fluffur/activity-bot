package stats

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/norm"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot                  *botapi.Bot
	translator           *i18n.Translator
	rules                *predicate.RuleChecker
	permissions          *predicate.PermissionChecker
	chatMemberRepository chatmember.Repository
	normRepository       norm.Repository
	statsRepository      Repository
}

func NewHandler(
	b *botapi.Bot,
	t *i18n.Translator,
	p *predicate.PermissionChecker,
	c *predicate.RuleChecker,
	cmr chatmember.Repository,
	nr norm.Repository,
	sr Repository,
) *Handler {
	return &Handler{b, t, c, p, cmr, nr, sr}
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
