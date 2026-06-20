package help

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/utils/tghtml"
	"fmt"

	"github.com/gotd/botapi"
)

func (h *Handler) Help(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("help chat ctx: %w", err)
	}

	args := i18n.HelpArgs(
		tghtml.Bold(tghtml.Link(h.commandsURL, h.translator.T(ch.Language, i18n.BotCommands))),
		tghtml.UserLink(h.developerUsername),
	)

	_, err = c.Reply(
		h.translator.TData(ch.Language, i18n.Help, args),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}
