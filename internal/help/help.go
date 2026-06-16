package help

import (
	"activity-bot/internal/i18n"

	"github.com/gotd/botapi"
)

func (h *Handler) Help(c *botapi.Context) error {
	_, err := c.Reply(
		h.translator.T("ru", i18n.Help, nil),
	)

	return err
}
