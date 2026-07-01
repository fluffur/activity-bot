package rest

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"

	"github.com/gotd/botapi"
)

const CategoryRest command.Category = "rest"

type Handler struct {
	bot               *botapi.Bot
	translator        *i18n.Translator
	permissions       *predicate.PermissionChecker
	rules             *predicate.RuleChecker
	chatMemberService *chatmember.Service
}

func NewHandler(
	b *botapi.Bot,
	t *i18n.Translator,
	p *predicate.PermissionChecker,
	rs *predicate.RuleChecker,
	cms *chatmember.Service,
) *Handler {
	return &Handler{
		bot:               b,
		translator:        t,
		chatMemberService: cms,
		permissions:       p,
		rules:             rs,
	}
}

func (h *Handler) Register(registry *command.Registry) {
	registry.AddCategory(CategoryRest)

	showRestDef := &command.ActionDef{
		Key:         "rest",
		Aliases:     []string{"рест"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusMember,
		Category:    CategoryRest,
		Description: i18n.Cmd.Rest.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
	}

	registry.Add(showRestDef)

	h.bot.OnMessage(
		h.ShowRest,
		predicate.Command(showRestDef.Key, showRestDef.Aliases...),
		predicate.NoArgs(),
		h.permissions.Require(showRestDef.Key, showRestDef.MinStatus),
	)
}
