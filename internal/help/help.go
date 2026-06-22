package help

import (
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"strings"

	"github.com/gotd/botapi"
)

func (h *Handler) Help(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("help chat ctx: %w", err)
	}
	lang := ch.Language

	groups := h.registry.ByCategory()

	var sb strings.Builder

	for _, cat := range command.Categories() {
		cmds := groups[cat]
		if len(cmds) == 0 {
			continue
		}
		sb.WriteString(
			tghtml.Bold(
				h.translator.T(lang, i18n.MessageID("category_"+string(cat))) + "\n",
			),
		)
		for _, c := range cmds {
			sb.WriteString("/" + c.Key + " — " +
				h.translator.T(lang, c.Description) + "\n")
			if len(c.Aliases) == 0 {
				continue
			}
			sb.WriteString(h.translator.TData(lang, i18n.Aliases, i18n.AliasesArgs(strings.Join(c.Aliases, " "))) + "\n")
		}
		sb.WriteString("\n")
	}

	_, err = c.Reply(
		sb.String(),
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
