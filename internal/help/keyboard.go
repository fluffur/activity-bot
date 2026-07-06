package help

import (
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"activity-bot/internal/utils/tghtml"
	"fmt"

	"github.com/gotd/botapi"
)

func (h *Handler) categoriesKeyboard(
	botUsername string,
	loc *i18n.Localizer,
) *botapi.InlineKeyboardMarkup {
	var rows [][]botapi.InlineKeyboardButton

	for _, category := range h.registry.Categories() {
		cmds := h.registry.ByCategory(category)

		hasCommands := false

		for _, cmd := range cmds {
			if cmd.ShowInHelp {
				hasCommands = true
				break
			}
		}

		if !hasCommands {
			continue
		}

		title := loc.T(
			i18n.MessageID("category."+string(category)),
			nil,
		)

		rows = append(rows,
			botapi.InlineRow(
				botapi.InlineButtonData(
					title,
					fmt.Sprintf("%s:%s", callbackHelpCategory, category),
				),
			),
		)
	}

	rows = append(rows,
		botapi.InlineRow(
			botapi.InlineButtonURL(
				loc.T(i18n.System.AddBotButton, nil),
				tghtml.StartGroupLink(botUsername),
			),
		),
	)

	return &botapi.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func (h *Handler) commandKeyboard(
	loc *i18n.Localizer,
	category command.Category,
	key string,
) *botapi.InlineKeyboardMarkup {
	var (
		rows [][]botapi.InlineKeyboardButton
		nav  []botapi.InlineKeyboardButton
	)

	h.registry.Find("")

	if prev := h.registry.Prev(category, key); prev != nil {
		nav = append(nav,
			botapi.InlineButtonData(
				"◀️",
				fmt.Sprintf(
					"%s:%s:%s",
					callbackHelpCommand,
					category,
					prev.Key,
				),
			),
		)
	}

	if next := h.registry.Next(category, key); next != nil {
		nav = append(nav,
			botapi.InlineButtonData(
				"▶️",
				fmt.Sprintf(
					"%s:%s:%s",
					callbackHelpCommand,
					category,
					next.Key,
				),
			),
		)
	}

	rows = append(rows,
		botapi.InlineRow(
			botapi.InlineButtonData(
				loc.T(i18n.Cmd.Help.ButtonCommands, nil),
				fmt.Sprintf("%s:%s", callbackHelpCategory, category),
			),
		),
	)

	if len(nav) > 0 {
		rows = append(rows, nav)
	}

	return &botapi.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func (h *Handler) commandsKeyboard(
	loc *i18n.Localizer,
	category command.Category,
) *botapi.InlineKeyboardMarkup {
	var rows [][]botapi.InlineKeyboardButton

	for _, cmd := range h.registry.ByCategory(category) {
		if !cmd.ShowInHelp {
			continue
		}

		rows = append(rows,
			botapi.InlineRow(
				botapi.InlineButtonData(
					"/"+cmd.Key,
					fmt.Sprintf(
						"%s:%s:%s",
						callbackHelpCommand,
						category,
						cmd.Key,
					),
				),
			),
		)
	}

	rows = append(rows,
		botapi.InlineRow(
			botapi.InlineButtonData(
				loc.T(i18n.Cmd.Help.ButtonBack, nil),
				callbackHelpCategories,
			),
		),
	)

	return &botapi.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}
