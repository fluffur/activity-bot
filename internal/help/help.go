package help

import (
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/predicate"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"log"
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

		categoryTitle := h.translator.T(lang, i18n.MessageID("category."+string(cat)))
		sb.WriteString(tghtml.Bold(categoryTitle) + "\n")

		for _, c := range cmds {
			desc := tghtml.Escape(h.translator.T(lang, c.Description))
			sb.WriteString(fmt.Sprintf("/%s — %s\n", c.Key, desc))

			hasDetails := len(c.Args) > 0 || len(c.Aliases) > 0 || len(c.Examples) > 0
			if hasDetails {
				sb.WriteString("<blockquote expandable>")

				var innerParts []string

				if len(c.Args) > 0 {
					syntax := buildSyntaxString(c.Key, c.Args, h.translator, lang)
					label := h.translator.T(lang, i18n.Cmd.Help.Syntax)
					innerParts = append(innerParts, tghtml.Italic(label)+" "+tghtml.Code(syntax))
				}

				if len(c.Aliases) > 0 {
					aliasesStr := strings.Join(c.Aliases, ", ")

					label := h.translator.TData(
						lang,
						i18n.Cmd.Help.AliasesLabel,
						i18n.CmdHelpAliasesLabelArgs(tghtml.Escape(aliasesStr)),
					)
					innerParts = append(innerParts, label)
				}

				if len(c.Examples) > 0 {
					label := h.translator.T(lang, i18n.Cmd.Help.Examples)
					exampleBlock := tghtml.Italic(label)
					for _, exID := range c.Examples {
						exampleText := h.translator.T(lang, exID)
						exampleBlock += "\n• " + tghtml.Code(exampleText)
					}
					innerParts = append(innerParts, exampleBlock)
				}

				sb.WriteString(strings.Join(innerParts, "\n"))
				sb.WriteString("</blockquote>\n")
			}
		}
		sb.WriteString("\n")
	}

	log.Println(sb.String())
	_, err = c.Reply(
		sb.String(),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(h.helpKeyboard(ch.Language)),
	)

	return err
}

func buildSyntaxString(commandName string, args []predicate.Arg, translator *i18n.Translator, lang string) string {
	var parts []string
	parts = append(parts, "/"+commandName)

	var argLabels = map[predicate.ArgType]i18n.MessageID{
		predicate.ArgTypeUser:     i18n.ArgType.User,
		predicate.ArgTypeNumber:   i18n.ArgType.Number,
		predicate.ArgTypeDuration: i18n.ArgType.Duration,
		predicate.ArgTypeDateTime: i18n.ArgType.Datetime,
		predicate.ArgTypeText:     i18n.ArgType.Text,
	}

	for _, arg := range args {
		var name string
		if msgID, ok := argLabels[arg.Type]; ok {
			name = translator.T(lang, msgID)
		} else {
			name = string(arg.Type)
		}

		if arg.Count == predicate.ArgCountVariadic {
			name += "..."
		}

		if arg.Optional {
			parts = append(parts, "["+name+"]")
		} else {
			parts = append(parts, "<"+name+">")
		}
	}

	return strings.Join(parts, " ")
}

func (h *Handler) helpKeyboard(lang string) *botapi.InlineKeyboardMarkup {
	return &botapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]botapi.InlineKeyboardButton{{
			{
				Text: h.translator.T(lang, i18n.System.AddBotButton),
				URL:  tghtml.StartGroupLink(h.bot.Self().Username),
			},
		}},
	}
}
