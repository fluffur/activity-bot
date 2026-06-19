package help

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/utils/tghtml"

	"github.com/gotd/botapi"
)

func (h *Handler) Help(c *botapi.Context) error {
	args := i18n.HelpArgs(
		tghtml.Bold(
			tghtml.Link(h.commandsURL, h.translator.T("ru", i18n.BotCommands, nil)),
		),
		tghtml.UserLink(h.developerUsername),
	)

	_, err := c.Reply(
		h.translator.T("ru", i18n.Help, args),
		botapi.WithParseMode(botapi.ParseModeHTML),
	)

	return err
}
