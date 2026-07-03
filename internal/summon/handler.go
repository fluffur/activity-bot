package summon

import (
	"activity-bot/internal/action"
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/predicate"
	"sync"

	fsm "github.com/fluffur/botapi-fsm"

	"github.com/gotd/botapi"
)

const CategorySummon command.Category = "summon"

type Handler struct {
	bot               *botapi.Bot
	permissions       *predicate.PermissionChecker
	chatService       *chat.Service
	chatMemberService *chatmember.Service
	activeSummons     sync.Map
	summonFSM         *fsm.Machine[State, StateData]
}

func NewHandler(
	b *botapi.Bot,
	p *predicate.PermissionChecker,
	chs *chat.Service,
	cms *chatmember.Service,
	summonFSM *fsm.Machine[State, StateData],
) *Handler {
	return &Handler{
		bot:               b,
		permissions:       p,
		chatService:       chs,
		chatMemberService: cms,
		activeSummons:     sync.Map{},
		summonFSM:         summonFSM,
	}
}

func (h *Handler) Register(registry *command.Registry) {
	summonDef := action.NewCommand(
		"summon",
		i18n.Cmd.Summon.Desc,
		CategorySummon,
		permission.StatusAdmin,
		option.WithAliases("call", "калл", "колл", "каллалл"),
	)

	unregDef := action.NewCommand(
		"unreg",
		i18n.Cmd.Summon.Reg.Desc,
		CategorySummon,
		permission.StatusMember,
		option.WithAliases("анрег"),
	)

	regDef := action.NewCommand(
		"reg",
		i18n.Cmd.Summon.Reg.Desc,
		CategorySummon,
		permission.StatusMember,
		option.WithAliases("рег"),
	)

	summonStyleDef := action.NewCommand(
		"summon_style",
		i18n.Cmd.Summon.Style.Desc,
		CategorySummon,
		permission.StatusMember,
		option.WithAliases("каллтип", "каллстиль", "калл тип", "калл стиль"),
	)

	summonStyleToggleDef := action.NewCallback(
		"toggle_summon_style",
		CategorySummon,
		permission.StatusSeniorAdmin,
		summonStyleDef,
	)

	registry.Add(summonDef)
	registry.Add(unregDef)
	registry.Add(regDef)
	registry.Add(summonStyleDef)
	registry.Add(summonStyleToggleDef)

	h.bot.OnMessage(h.SummonStyle,
		predicate.Chat(),
		predicate.Command(summonStyleDef.Key, summonStyleDef.Aliases...),
		predicate.NoArgs(),
		h.permissions.Require(summonStyleDef.Key, summonStyleDef.MinStatus),
	)

	h.bot.OnMessage(h.SummonAll,
		predicate.Chat(),
		predicate.Command(summonDef.Key, summonDef.Aliases...),
		h.permissions.Require(summonDef.Key, summonDef.MinStatus),
	)

	h.bot.OnCallbackQuery(
		h.ConfirmSummon,
		predicate.Chat(),
		botapi.CallbackData("summon:confirm"),
		h.permissions.Require(summonDef.Key, summonDef.MinStatus),
	)

	h.bot.OnCallbackQuery(
		h.CancelSummon,
		predicate.Chat(),
		botapi.CallbackData("summon:cancel"),
		h.permissions.Require(summonDef.Key, summonDef.MinStatus),
	)

	h.bot.OnCallbackQuery(
		h.ConfirmSummonDontAsk,
		predicate.Chat(),
		botapi.CallbackData("summon:confirm_dont_ask"),
		h.permissions.Require(summonDef.Key, summonDef.MinStatus),
	)

	h.bot.OnMessage(h.Unreg,
		predicate.Chat(),
		predicate.Command(unregDef.Key, unregDef.Aliases...),
		predicate.NoArgs(),
		h.permissions.Require(unregDef.Key, unregDef.MinStatus),
	)

	h.bot.OnMessage(h.Reg,
		predicate.Chat(),
		predicate.Command(regDef.Key, regDef.Aliases...),
		predicate.NoArgs(),
		h.permissions.Require(regDef.Key, regDef.MinStatus),
	)

	h.bot.OnCallbackQuery(h.ToggleMentionEmoji,
		predicate.Chat(),
		botapi.CallbackData("summon:style:emoji"),
		h.permissions.Require(summonStyleToggleDef.Key, summonStyleToggleDef.MinStatus),
	)
	h.bot.OnCallbackQuery(h.ToggleMentionName,
		predicate.Chat(),
		botapi.CallbackData("summon:style:name"),
		h.permissions.Require(summonStyleToggleDef.Key, summonStyleToggleDef.MinStatus),
	)
	h.bot.OnCallbackQuery(h.ToggleMentionRole,
		predicate.Chat(),
		botapi.CallbackData("summon:style:role"),
		h.permissions.Require(summonStyleToggleDef.Key, summonStyleToggleDef.MinStatus),
	)
}
