package help

import (
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"log"

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
	log.Println(h.translator.TData(ch.Language, i18n.Help, args))

	_, err = c.Reply(
		h.translator.TData(ch.Language, i18n.Help, args),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(h.helpKeyboard(ch.Language)),
	)

	return err
}

func (h *Handler) helpKeyboard(lang string) *botapi.InlineKeyboardMarkup {
	return &botapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]botapi.InlineKeyboardButton{{
			{
				Text: h.translator.T(lang, i18n.AddBotButton),
				URL:  tghtml.StartGroupLink(h.bot.Self().Username),
			},
		}},
	}
}
