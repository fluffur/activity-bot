package help

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/i18n"
	"activity-bot/internal/predicate"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"log"
	"strings"

	"github.com/gotd/botapi"
)

func (h *Handler) Help(c *botapi.Context) error {
	loc := cctx.MustLocalizer(c)

	groups := h.registry.ByCategory()
	var sb strings.Builder

	for _, cat := range h.registry.Categories() {
		cmds := groups[cat]
		if len(cmds) == 0 {
			continue
		}

		categoryTitle := loc.T(i18n.MessageID("category."+string(cat)), nil)
		sb.WriteString(tghtml.Bold(categoryTitle) + "\n")

		for _, cmd := range cmds {
			if !cmd.ShowInHelp {
				continue
			}
			desc := tghtml.Escape(loc.T(cmd.Description, nil))
			sb.WriteString(fmt.Sprintf("/%s — %s\n", cmd.Key, desc))

			hasDetails := len(cmd.Rules) > 0 || len(cmd.Aliases) > 0 || len(cmd.Examples) > 0
			if hasDetails {
				sb.WriteString("<blockquote expandable>")

				var innerParts []string

				if len(cmd.Rules) > 0 {
					syntax := buildSyntaxString(cmd.Key, cmd.Rules, loc)
					label := loc.T(i18n.Cmd.Help.Syntax, nil)
					innerParts = append(innerParts, tghtml.Italic(label)+" "+tghtml.Code(syntax))
				}

				if len(cmd.Aliases) > 0 {
					aliasesStr := strings.Join(cmd.Aliases, ", ")

					label := loc.T(
						i18n.Cmd.Help.AliasesLabel,
						i18n.CmdHelpAliasesLabelData{
							Aliases: tghtml.Escape(aliasesStr),
						},
					)
					innerParts = append(innerParts, label)
				}

				if len(cmd.Examples) > 0 {
					label := loc.T(i18n.Cmd.Help.Examples, nil)
					exampleBlock := tghtml.Italic(label)
					for _, exID := range cmd.Examples {
						exampleText := loc.T(exID, nil)
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
	_, err := c.Reply(
		sb.String(),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.WithReplyMarkup(h.helpKeyboard(loc)),
	)

	return err
}

func buildSyntaxString(commandName string, args []predicate.Rule, loc *i18n.Localizer) string {
	var parts []string
	parts = append(parts, "/"+commandName)

	var argLabels = map[predicate.RuleType]i18n.MessageID{
		predicate.RuleUser:     i18n.ArgType.User,
		predicate.RuleNumber:   i18n.ArgType.Number,
		predicate.RuleDuration: i18n.ArgType.Duration,
		predicate.RuleDateTime: i18n.ArgType.Datetime,
		predicate.RuleText:     i18n.ArgType.Text,
	}

	for _, arg := range args {
		var name string
		if msgID, ok := argLabels[arg.Type]; ok {
			name = loc.T(msgID, nil)
		} else {
			name = string(arg.Type)
		}

		if arg.Count == predicate.RuleVariadic {
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

func (h *Handler) helpKeyboard(loc *i18n.Localizer) *botapi.InlineKeyboardMarkup {
	return &botapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]botapi.InlineKeyboardButton{{
			{
				Text: loc.T(i18n.System.AddBotButton, nil),
				URL:  tghtml.StartGroupLink(h.bot.Self().Username),
			},
		}},
	}
}
