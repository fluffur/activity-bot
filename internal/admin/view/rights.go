package view

import (
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"strings"

	"github.com/gotd/td/tg"
)

func FormatCategories() string {
	return "⚖️ <b>Настройка прав доступа команд</b>\n\nВыберите категорию команд для настройки:"
}

func GetCategoriesKeyboard() tg.ReplyMarkupClass {
	categories := command.Categories()

	var rows []tg.KeyboardButtonRow
	for _, cat := range categories {
		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonCallback{
					Text: string(cat),
					Data: []byte("rights_cat:" + string(cat)),
				},
			},
		})
	}
	return &tg.ReplyInlineMarkup{Rows: rows}
}

func FormatCommandsByCategory(category command.Category, commands []*command.Command, perms map[string]model.Status) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Категория: %s\n\n", category))

	for _, c := range commands {
		if c.Category() != category {
			continue
		}

		status := c.RequiredStatus()
		if s, ok := perms[c.Name()]; ok {
			status = s
		}

		aliases := append(c.Aliases(), c.Name())

		formattedAliases := make([]string, len(aliases))
		for i, a := range aliases {
			formattedAliases[i] = "<code>" + a + "</code>"
		}

		sb.WriteString(fmt.Sprintf("%s %s (%s)\n",
			helpers.StatusEmoji(status),
			c.Description(),
			strings.Join(formattedAliases, ", "),
		))
	}

	sb.WriteString("\nВыберите команду для изменения прав:")
	return sb.String()
}

func GetCommandsByCategoryKeyboard(category command.Category, commands []*command.Command, perms map[string]model.Status) tg.ReplyMarkupClass {
	var rows []tg.KeyboardButtonRow
	for _, c := range commands {
		if c.Category() != category {
			continue
		}

		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonCallback{
					Text: c.Description(),
					Data: []byte("rights_edit:" + c.Name()),
				},
			},
		})
	}

	rows = append(rows, tg.KeyboardButtonRow{
		Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButtonCallback{
				Text: "К категориям",
				Data: []byte("rights_list"),
			},
		},
	})

	return &tg.ReplyInlineMarkup{Rows: rows}
}

func FormatEditCommandRights(key string, currentStatus model.Status, commands []*command.Command) string {
	var config *command.Command
	for _, c := range commands {
		if c.Name() == key {
			config = c
			break
		}
	}

	if config == nil {
		return "Ошибка: команда не найдена"
	}

	aliases := append(config.Aliases(), config.Name())
	formattedAliases := make([]string, len(aliases))
	for i, a := range aliases {
		formattedAliases[i] = "<code>" + a + "</code>"
	}

	return fmt.Sprintf("⚙️ %s\nСинонимы: %s\n\nТекущий уровень: %s %s\n\nВыберите новый уровень доступа:",
		config.Description(),
		strings.Join(formattedAliases, ", "),
		helpers.StatusEmoji(currentStatus),
		currentStatus.String(),
	)
}

func GetEditRightsKeyboard(key string, currentStatus model.Status, commands []*command.Command) tg.ReplyMarkupClass {
	statuses := []model.Status{
		model.StatusMember,
		model.StatusModerator,
		model.StatusAdmin,
		model.StatusSeniorAdmin,
		model.StatusCoOwner,
		model.StatusOwner,
	}

	var rows []tg.KeyboardButtonRow
	for _, s := range statuses {
		text := s.Title()
		if s == currentStatus {
			text = "✅ " + text
		}
		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonCallback{
					Text: text,
					Data: []byte(fmt.Sprintf("rights_set:%s:%d", key, s)),
				},
			},
		})
	}

	var backCategory string
	for _, c := range commands {
		if c.Name() == key {
			backCategory = string(c.Category())
			break
		}
	}

	rows = append(rows, tg.KeyboardButtonRow{
		Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButtonCallback{
				Text: "Назад",
				Data: []byte("rights_cat:" + backCategory),
			},
		},
	})

	return &tg.ReplyInlineMarkup{Rows: rows}
}
