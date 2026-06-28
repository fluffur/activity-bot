package summon

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"fmt"

	"github.com/gotd/botapi"
)

func (h *Handler) Unreg(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("exclude from summon chat: %w", err)

	}

	cm, err := cctx.ChatMember(c.Context)
	if err != nil {
		return fmt.Errorf("exclude from summon cm: %w", err)
	}

	if err := h.chatMemberService.SetExcludeFromSummon(c.Context, ch.ID, cm.User.ID, true); err != nil {
		return fmt.Errorf("exclude from summon set: %w", err)
	}

	_, err = c.Reply(
		h.translator.T(ch.Lang, i18n.Cmd.Unreg.Set),
	)

	return err
}

func (h *Handler) Reg(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("include to summon chat: %w", err)

	}

	cm, err := cctx.ChatMember(c.Context)
	if err != nil {
		return fmt.Errorf("include to summon cm: %w", err)
	}

	if err := h.chatMemberService.SetExcludeFromSummon(c.Context, ch.ID, cm.User.ID, false); err != nil {
		return fmt.Errorf("include to summon set: %w", err)
	}

	_, err = c.Reply(
		h.translator.T(ch.Lang, i18n.Cmd.Reg.Set),
	)

	return err
}
