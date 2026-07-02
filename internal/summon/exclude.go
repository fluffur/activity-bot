package summon

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/i18n"
	"fmt"

	"github.com/gotd/botapi"
)

func (h *Handler) Unreg(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)
	cm := cctx.MustChatMember(c)

	if err := h.chatMemberService.SetExcludeFromSummon(c, ch.ID, cm.ID(), true); err != nil {
		return fmt.Errorf("exclude from summon set: %w", err)
	}

	_, err := c.Reply(
		loc.T(i18n.Cmd.Summon.Unreg.Set, nil),
	)

	return err
}

func (h *Handler) Reg(c *botapi.Context) error {
	ch := cctx.MustChat(c)
	loc := cctx.MustLocalizer(c)
	cm := cctx.MustChatMember(c)

	if err := h.chatMemberService.SetExcludeFromSummon(c, ch.ID, cm.ID(), false); err != nil {
		return fmt.Errorf("include to summon set: %w", err)
	}

	_, err := c.Reply(
		loc.T(i18n.Cmd.Summon.Reg.Set, nil),
	)

	return err
}
