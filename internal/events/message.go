package events

import (
	"activity-bot/internal/middleware/cctx"
	"fmt"
	"log"

	"github.com/gotd/botapi"
)

func (h *Handler) Message(c *botapi.Context) error {
	sender := c.Sender()
	message := c.Message()

	if message == nil || sender == nil {
		return nil
	}

	if err := h.repository.Save(
		c.Context,
		message.Chat.ID,
		sender.ID,
		int64(message.MessageID),
	); err != nil {
		return fmt.Errorf("save message: %w", err)
	}
	m, err := cctx.ChatMember(c.Context)
	if err != nil {
		return err
	}
	log.Printf("%+v\n", m)
	return nil
}
