package moderation

import (
	"activity-bot/internal/cctx"
	"fmt"
	"time"

	"github.com/gotd/botapi"
)

func (h *Handler) Ban(c *botapi.Context) error {
	self := cctx.MustChatMember(c)
	args := cctx.MustArgs(c)
	target, ok := args.User()

	if !ok || self.ID() == target.ID() {
		return nil
	}

	until, ok := args.Until()
	if !ok {
		until = time.Time{}
	}

	ch := cctx.MustChat(c)

	if err := h.bot.BanChatMember(c, botapi.ID(ch.ID), target.ID(), botapi.WithBanUntil(int(until.Unix()))); err != nil {
		return fmt.Errorf("ban: %w", err)
	}

	return nil
}
