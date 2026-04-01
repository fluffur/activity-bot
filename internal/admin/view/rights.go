package view

import (
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func FormatCategories() string {
	return "⚖️ <b>Настройка прав доступа команд</b>\n\nВыберите категорию команд для настройки:"
}

func GetCategoriesKeyboard() gotgbot.InlineKeyboardMarkup {
	categories := []command.Category{
		command.CategoryStats,
		command.CategoryModeration,
		command.CategoryCall,
		command.CategorySettings,
	}

	var rows [][]gotgbot.InlineKeyboardButton
	for _, cat := range categories {
		rows = append(rows, []gotgbot.InlineKeyboardButton{
			{Text: string(cat), CallbackData: "rights_cat:" + string(cat)},
		})
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
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

		aliases := c.Aliases()
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

func GetCommandsByCategoryKeyboard(category command.Category, commands []*command.Command, perms map[string]model.Status) gotgbot.InlineKeyboardMarkup {
	var rows [][]gotgbot.InlineKeyboardButton
	for _, c := range commands {
		if c.Category() != category {
			continue
		}

		status := c.RequiredStatus()
		if s, ok := perms[c.Name()]; ok {
			status = s
		}

		rows = append(rows, []gotgbot.InlineKeyboardButton{
			{Text: c.Description(), CallbackData: "rights_edit:" + c.Name(), IconCustomEmojiId: helpers.StatusEmojiID(status)},
		})
	}

	rows = append(rows, []gotgbot.InlineKeyboardButton{
		{Text: "К категориям", CallbackData: "rights_list", Style: "danger"},
	})

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
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

	aliases := config.Aliases()
	formattedAliases := make([]string, len(aliases))
	for i, a := range aliases {
		formattedAliases[i] = "<code>" + a + "</code>"
	}

	return fmt.Sprintf("⚙️ %s\nАлиасы: %s\n\nТекущий уровень: %s %s\n\nВыберите новый уровень доступа:",
		config.Description(),
		strings.Join(formattedAliases, ", "),
		helpers.StatusEmoji(currentStatus),
		currentStatus.String(),
	)
}

func GetEditRightsKeyboard(key string, currentStatus model.Status, commands []*command.Command) gotgbot.InlineKeyboardMarkup {
	statuses := []model.Status{
		model.StatusMember,
		model.StatusModerator,
		model.StatusAdmin,
		model.StatusSeniorAdmin,
		model.StatusCoOwner,
		model.StatusOwner,
	}

	var rows [][]gotgbot.InlineKeyboardButton
	var currentRow []gotgbot.InlineKeyboardButton
	for _, s := range statuses {
		style := ""
		if s == currentStatus {
			style = "primary"
		}
		currentRow = append(currentRow, gotgbot.InlineKeyboardButton{
			Text:              s.Title(),
			CallbackData:      fmt.Sprintf("rights_set:%s:%d", key, s),
			Style:             style,
			IconCustomEmojiId: helpers.StatusEmojiID(s),
		})
		//if (i+1)%3 == 0 {
		rows = append(rows, currentRow)
		currentRow = []gotgbot.InlineKeyboardButton{}
		//}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	var backCategory string
	for _, c := range commands {
		if c.Name() == key {
			backCategory = string(c.Category())
			break
		}
	}

	rows = append(rows, []gotgbot.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "rights_cat:" + backCategory, Style: "danger"},
	})

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}
