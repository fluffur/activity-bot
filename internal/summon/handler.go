package summon

import (
	"activity-bot/internal/action"
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/option"
	"activity-bot/internal/permission"
	"activity-bot/internal/rule"
	"sync"

	fsm "github.com/fluffur/botapi-fsm"

	"github.com/gotd/botapi"
)

const CategorySummon command.Category = "summon"

const (
	callbackSummonStyle          = "summon:style:"
	callbackSummonConfirm        = "summon:confirm"
	callbackSummonCancel         = "summon:cancel"
	callbackSummonConfirmDontAsk = "summon:confirm_dont_ask"
)

type Handler struct {
	chatService       *chat.Service
	chatMemberService *chatmember.Service
	activeSummons     sync.Map
	summonFSM         *fsm.Machine[State, StateData]
}

func NewHandler(
	chs *chat.Service,
	cms *chatmember.Service,
	summonFSM *fsm.Machine[State, StateData],
) *Handler {
	return &Handler{
		chatService:       chs,
		chatMemberService: cms,
		activeSummons:     sync.Map{},
		summonFSM:         summonFSM,
	}
}
func (h *Handler) Actions() []*command.Action {
	return []*command.Action{

		action.NewCommand(
			"summonstyle",
			h.SummonStyle,
			i18n.Cmd.Summon.Style.Desc,
			CategorySummon,
			option.WithAliases(
				"каллтип",
				"каллстиль",
				"калл тип",
				"калл стиль",
			),
		),
		action.NewCommand(
			"summon",
			h.SummonAll,
			i18n.Cmd.Summon.Desc,
			CategorySummon,
			option.WithPermission(permission.StatusAdmin),
			option.WithRules(rule.Text().Optional()),
			option.WithAliases("call", "калл", "колл", "каллалл"),
		),

		action.NewCommand(
			"unreg",
			h.Unreg,
			i18n.Cmd.Summon.Unreg.Desc,
			CategorySummon,
			option.WithAliases("анрег"),
		),

		action.NewCommand(
			"reg",
			h.Reg,
			i18n.Cmd.Summon.Reg.Desc,
			CategorySummon,
			option.WithAliases("рег"),
		),

		action.NewCallbackPrefix(
			"toggle_summon_style",
			callbackSummonStyle,
			h.ToggleSummonStyle,
			CategorySummon,
			option.WithPermission(permission.StatusSeniorAdmin),
		),

		action.NewCallback(
			"confirm_summon",
			callbackSummonConfirm,
			h.ConfirmSummon,
			CategorySummon,
			option.WithPredicates(h.summonFSM.State(StateAwaitConfirmation)),
			option.WithPermission(permission.StatusAdmin),
		),

		action.NewCallback(
			"cancel_summon",
			callbackSummonCancel,
			h.CancelSummon,
			CategorySummon,
			option.WithPredicates(h.summonFSM.State(StateAwaitConfirmation)),
			option.WithPermission(permission.StatusAdmin),
		),

		action.NewCallback(
			"confirm_summon_dont_ask",
			callbackSummonConfirmDontAsk,
			h.ConfirmSummonDontAsk,
			CategorySummon,
			option.WithPredicates(h.summonFSM.State(StateAwaitConfirmation)),
			option.WithPermission(permission.StatusAdmin),
		),
		action.NewCommand(
			"summonuser",
			h.SummonSpecific,
			i18n.Cmd.SummonSpecific.Desc,
			CategorySummon,
			option.WithAliases("позвать", "призвать"),
			option.WithRules(rule.User(), rule.Text().Optional()),
			option.WithPermission(permission.StatusAdmin),
		),
	}
}

func (h *Handler) SummonSpecific(c *botapi.Context) error {
	args := cctx.MustArgs(c)
	users := args.Users
	if len(users) == 0 {
		return nil
	}
	msg := c.Message()
	if msg == nil {
		return nil
	}
	text, _ := args.Text()
	ch := cctx.MustChat(c)

	return h.Summon(c, text, msg.MessageID, ch, users)
}
