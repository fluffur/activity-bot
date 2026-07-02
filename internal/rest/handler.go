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
	bot         *botapi.Bot
	permissions *predicate.PermissionChecker
	rules       *predicate.RuleChecker

	chatMemberSerivce *chatmember.Service
	service           *Service
}

func NewHandler(
	b *botapi.Bot,
	p *predicate.PermissionChecker,
	rs *predicate.RuleChecker,
	rsrv *Service,
	cms *chatmember.Service,
) *Handler {
	return &Handler{
		bot:               b,
		permissions:       p,
		rules:             rs,
		service:           rsrv,
		chatMemberSerivce: cms,
	}
}

func (h *Handler) Register(registry *command.Registry) {
	registry.AddCategory(CategoryRest)

	showRestDef := &command.ActionDef{
		Key:         "rest",
		Aliases:     []string{"рест", "мой рест"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusMember,
		Category:    CategoryRest,
		Description: i18n.Cmd.Rest.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleUser, Optional: true, Count: 1},
		},
	}

	setRestDef := &command.ActionDef{
		Key:         "set_rest",
		Aliases:     []string{"рест", "+рест", "установить рест"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusModerator,
		Category:    CategoryRest,
		Description: i18n.Cmd.Rest.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleUser, Optional: true, Count: 1},
			{Type: predicate.RuleDurationOrDateTime, Optional: false, Count: 1},
			{Type: predicate.RuleText, Optional: true, Count: 1},
		},
	}

	endRestDef := &command.ActionDef{
		Key:         "end_rest",
		Aliases:     []string{"-рест", "завершить рест"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusModerator,
		Category:    CategoryRest,
		Description: i18n.Cmd.EndRest.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleUser, Optional: true, Count: 1},
		},
	}

	allRestDef := &command.ActionDef{
		Key:         "rests",
		Aliases:     []string{"ресты", "все ресты"},
		Trigger:     command.TriggerCommand,
		MinStatus:   chatmember.StatusModerator,
		Category:    CategoryRest,
		Description: i18n.Cmd.Rests.Desc,
		Scope:       command.ScopeGroup,
		ShowInHelp:  true,
		Rules: []predicate.Rule{
			{Type: predicate.RuleUser, Optional: true, Count: 1},
		},
	}

	approveRequestDef := &command.ActionDef{
		Key:        "approve:",
		Trigger:    command.TriggerCallback,
		Parent:     setRestDef,
		MinStatus:  chatmember.StatusModerator,
		Category:   CategoryRest,
		Scope:      command.ScopeGroup,
		ShowInHelp: false,
	}

	rejectRequestDef := &command.ActionDef{
		Key:        "reject:",
		Trigger:    command.TriggerCallback,
		Parent:     setRestDef,
		MinStatus:  chatmember.StatusModerator,
		Category:   CategoryRest,
		Scope:      command.ScopeGroup,
		ShowInHelp: false,
	}

	removeRequestDef := &command.ActionDef{
		Key:       "rests_delete",
		Aliases:   []string{"-рест", "-ресты", "удалить рест"},
		Trigger:   command.TriggerCommand,
		MinStatus: chatmember.StatusModerator,
		Category:  CategoryRest,
		Scope:     command.ScopeGroup,
		Rules: []predicate.Rule{
			{Type: predicate.RuleUser, Optional: true, Count: 1},
			{Type: predicate.RuleNumber, Optional: false, Count: 1},
		},
		ShowInHelp: true,
	}

	registry.Add(showRestDef)
	registry.Add(setRestDef)
	registry.Add(endRestDef)
	registry.Add(allRestDef)
	registry.Add(approveRequestDef)
	registry.Add(rejectRequestDef)
	registry.Add(removeRequestDef)

	h.bot.OnMessage(
		h.ShowRest,
		predicate.Command(showRestDef.Key, showRestDef.Aliases...),
		h.rules.With(showRestDef.Rules...),
		h.permissions.Require(showRestDef.Key, showRestDef.MinStatus),
	)

	h.bot.OnMessage(
		h.SetRest,
		predicate.Command(setRestDef.Key, setRestDef.Aliases...),
		h.rules.With(setRestDef.Rules...),
		h.permissions.Pass(setRestDef.Key, setRestDef.MinStatus),
	)

	h.bot.OnMessage(
		h.EndRest,
		predicate.Command(endRestDef.Key, endRestDef.Aliases...),
		h.rules.With(endRestDef.Rules...),
		h.permissions.Pass(endRestDef.Key, endRestDef.MinStatus),
	)

	h.bot.OnCallbackQuery(
		h.ApproveRestRequest,
		botapi.CallbackPrefix(approveRequestDef.Key),
		h.permissions.Require(approveRequestDef.Key, approveRequestDef.MinStatus),
	)

	h.bot.OnCallbackQuery(
		h.RejectRestRequest,
		botapi.CallbackPrefix(rejectRequestDef.Key),
		h.permissions.Pass(rejectRequestDef.Key, rejectRequestDef.MinStatus),
	)

	h.bot.OnMessage(
		h.AllUserRests,
		predicate.Command(allRestDef.Key, allRestDef.Aliases...),
		h.rules.With(allRestDef.Rules...),
		h.permissions.Require(allRestDef.Key, allRestDef.MinStatus),
	)

	h.bot.OnMessage(
		h.RemoveRestRequest,
		predicate.Command(removeRequestDef.Key, removeRequestDef.Aliases...),
		h.rules.With(removeRequestDef.Rules...),
		h.permissions.Require(removeRequestDef.Key, removeRequestDef.MinStatus),
	)

}
