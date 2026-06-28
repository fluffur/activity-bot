package summon

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"
	"sync"

	"github.com/gotd/botapi"
)

type Handler struct {
	bot               *botapi.Bot
	translator        *i18n.Translator
	permissions       *predicate.PermissionChecker
	chatMemberService *chatmember.Service

	activeSummons sync.Map
}

func NewHandler(b *botapi.Bot, t *i18n.Translator, p *predicate.PermissionChecker, cms *chatmember.Service) *Handler {
	return &Handler{b, t, p, cms, sync.Map{}}
}

func (h *Handler) Register(registry *command.Registry) {
	summonDef := &command.ActionDef{
		Key:         "summon",
		Aliases:     []string{"call", "калл", "колл", "каллалл"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusAdmin,
		Category:    command.CategorySummon,
		Description: i18n.Cmd.Summon.Desc,
		Scope:       command.ScopeGroup,
		Examples:    []i18n.MessageID{},
		ShowInHelp:  true,
	}

	unregDef := &command.ActionDef{
		Key:         "unreg",
		Aliases:     []string{"анрег"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusMember,
		Category:    command.CategorySummon,
		Description: i18n.Cmd.Unreg.Desc,
		Examples:    []i18n.MessageID{},
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
	}

	regDef := &command.ActionDef{
		Key:         "reg",
		Aliases:     []string{"рег"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusMember,
		Category:    command.CategorySummon,
		Description: i18n.Cmd.Reg.Desc,
		Examples:    []i18n.MessageID{},
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
	}

	registry.Add(summonDef)
	registry.Add(unregDef)
	registry.Add(regDef)

	h.bot.OnMessage(h.SummonAll,
		predicate.Command(summonDef.Key, summonDef.Aliases...),
		h.permissions.Require(summonDef.Key, summonDef.MinStatus),
	)

	h.bot.OnMessage(h.Unreg,
		predicate.Command(unregDef.Key, unregDef.Aliases...),
		predicate.NoArgs(),
		h.permissions.Require(unregDef.Key, unregDef.MinStatus),
	)

	h.bot.OnMessage(h.Reg,
		predicate.Command(regDef.Key, regDef.Aliases...),
		predicate.NoArgs(),
		h.permissions.Require(regDef.Key, regDef.MinStatus),
	)
}
