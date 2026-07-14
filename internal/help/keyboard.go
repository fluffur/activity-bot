package help

import (
	"activity-bot/internal/command"
	"activity-bot/internal/i18n"
	"fmt"

	"github.com/gotd/botapi"
)

func (h *Handler) categoriesKeyboard(loc *i18n.Localizer) *botapi.InlineKeyboardMarkup {
	var (
		rows [][]botapi.InlineKeyboardButton
		row  []botapi.InlineKeyboardButton
	)

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

		row = append(row,
			botapi.InlineButtonData(
				title,
				fmt.Sprintf(
					"%s:%s:%d",
					callbackHelpCategory,
					category,
					0,
				),
			),
		)

		// две кнопки в строке
		if len(row) == 2 {
			rows = append(rows, row)
			row = nil
		}
	}

	// если осталась одна кнопка
	if len(row) > 0 {
		rows = append(rows, row)
	}

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

	if len(nav) > 0 {
		rows = append(rows, nav)
	}

	rows = append(rows,
		botapi.InlineRow(
			botapi.InlineButtonData(
				loc.T(i18n.Cmd.Help.ButtonCommands, nil),
				fmt.Sprintf("%s:%s", callbackHelpCategory, category),
			),
		),
	)

	return &botapi.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}

func (h *Handler) commandsKeyboard(
	loc *i18n.Localizer,
	category command.Category,
	page int,
) *botapi.InlineKeyboardMarkup {
	var rows [][]botapi.InlineKeyboardButton

	cmds := make([]*command.Action, 0)

	for _, cmd := range h.registry.ByCategory(category) {
		if cmd.ShowInHelp {
			cmds = append(cmds, cmd)
		}
	}
	if len(cmds) == 0 {
		return &botapi.InlineKeyboardMarkup{}
	}

	totalPages := (len(cmds) + commandsPerPage - 1) / commandsPerPage

	if page < 0 {
		page = 0
	}
	if page >= totalPages {
		page = totalPages - 1
	}

	start := page * commandsPerPage
	end := start + commandsPerPage

	if end > len(cmds) {
		end = len(cmds)
	}

	for _, cmd := range cmds[start:end] {
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

	if totalPages > 1 {
		var nav []botapi.InlineKeyboardButton

		if page > 0 {
			nav = append(nav,
				botapi.InlineButtonData(
					"◀️",
					fmt.Sprintf(
						"%s:%s:%d",
						callbackHelpCategory,
						category,
						page-1,
					),
				),
			)
		}

		nav = append(nav,
			botapi.InlineButtonData(
				fmt.Sprintf("%d/%d", page+1, totalPages),
				"ignore",
			),
		)

		if page+1 < totalPages {
			nav = append(nav,
				botapi.InlineButtonData(
					"▶️",
					fmt.Sprintf(
						"%s:%s:%d",
						callbackHelpCategory,
						category,
						page+1,
					),
				),
			)
		}

		rows = append(rows, nav)
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
