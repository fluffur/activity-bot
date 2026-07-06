package help

import (
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/rule"
	"activity-bot/internal/utils/tghtml"
	"strings"
)

func (h *Handler) renderCategories(loc *i18n.Localizer) string {
	var sb strings.Builder

	sb.WriteString("📖 ")
	sb.WriteString(tghtml.Bold(loc.T(i18n.Cmd.Help.Desc, nil)))
	sb.WriteString("\n")
	sb.WriteString(loc.T(i18n.Cmd.Help.Body, nil))
	sb.WriteString("\n\n")
	sb.WriteString(loc.T(i18n.Cmd.Help.ChooseCategory, nil))

	return sb.String()
}

func (h *Handler) renderCategory(
	loc *i18n.Localizer,
	category command.Category,
) string {
	var sb strings.Builder

	title := loc.T(i18n.MessageID("category."+string(category)), nil)

	sb.WriteString("📂 ")
	sb.WriteString(tghtml.Bold(title))
	sb.WriteString("\n\n")

	for _, cmd := range h.registry.ByCategory(category) {
		if !cmd.ShowInHelp {
			continue
		}

		sb.WriteString("• ")
		sb.WriteString(tghtml.Code("/" + cmd.Key))
		sb.WriteString(" — ")
		sb.WriteString(
			tghtml.Escape(
				loc.T(cmd.Description, nil),
			),
		)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (h *Handler) renderCommand(
	loc *i18n.Localizer,
	cmd *command.ActionDef,
) string {
	var sb strings.Builder

	sb.WriteString(tghtml.Bold("/" + cmd.Key))
	sb.WriteString("\n\n")

	sb.WriteString(
		tghtml.Escape(
			loc.T(cmd.Description, nil),
		),
	)

	sb.WriteString("\n\n")

	if trigger, ok := cmd.Trigger.(*command.CommandTrigger); ok {
		rules := trigger.Rules
		if len(rules) > 0 {
			sb.WriteString(tghtml.Bold(loc.T(i18n.Cmd.Help.Syntax, nil)))
			sb.WriteString("\n")

			sb.WriteString(
				tghtml.Code(
					buildSyntaxString(cmd.Key, rules, loc),
				),
			)

			sb.WriteString("\n\n")
		}

		aliases := trigger.Aliases
		if len(aliases) > 0 {
			sb.WriteString(tghtml.Bold(loc.T(i18n.Cmd.Help.AliasesLabel, nil)))
			sb.WriteString("\n")

			for _, alias := range aliases {
				sb.WriteString("• ")
				sb.WriteString(tghtml.Code(alias))
				sb.WriteString("\n")
			}

			sb.WriteString("\n")
		}
	}

	if len(cmd.Examples) > 0 {
		sb.WriteString(tghtml.Bold(loc.T(i18n.Cmd.Help.Examples, nil)))
		sb.WriteString("\n")

		for _, ex := range cmd.Examples {
			sb.WriteString("• ")
			sb.WriteString(
				tghtml.Code(
					loc.T(ex, nil),
				),
			)
			sb.WriteString("\n")
		}
	}

	return strings.TrimSpace(sb.String())
}

func buildSyntaxString(
	commandName string,
	args []rule.Rule,
	loc *i18n.Localizer,
) string {
	var parts []string

	parts = append(parts, "/"+commandName)

	labels := map[rule.RuleType]i18n.MessageID{
		rule.RuleUser:     i18n.ArgType.User,
		rule.RuleNumber:   i18n.ArgType.Number,
		rule.RuleDuration: i18n.ArgType.Duration,
		rule.RuleDateTime: i18n.ArgType.Datetime,
		rule.RuleText:     i18n.ArgType.Text,
	}

	for _, arg := range args {
		name := string(arg.Type)

		if id, ok := labels[arg.Type]; ok {
			name = loc.T(id, nil)
		}

		if arg.CountArgs == rule.RuleVariadic {
			name += "…"
		}

		if arg.IsOptional {
			name = "[" + name + "]"
		} else {
			name = "<" + name + ">"
		}

		parts = append(parts, name)
	}

	return strings.Join(parts, " ")
}
