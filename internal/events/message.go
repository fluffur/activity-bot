package events

import (
	"fmt"
	"log"

	"github.com/gotd/botapi"
)

func (h *Handler) Message(c *botapi.Context) error {
	sender := c.Sender()
	message := c.Message()
	log.Println("No message to send", message, sender)

	if message == nil || sender == nil {
		return nil
	}

	_, err := c.Reply(
		fmt.Sprintf("sender: %d\nmessage: %s\nchat: %d", sender.ID, message.Text, message.Chat.ID),
	)
	return err
}
