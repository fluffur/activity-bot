package events

import (
	"fmt"

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

	return nil
}
