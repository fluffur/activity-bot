package view

import (
	"activity-bot/internal/cmd"
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
	categories := []cmd.Category{
		cmd.CategoryStats,
		cmd.CategoryModeration,
		cmd.CategoryCall,
		cmd.CategorySettings,
	}

	var rows [][]gotgbot.InlineKeyboardButton
	for _, cat := range categories {
		rows = append(rows, []gotgbot.InlineKeyboardButton{
			{Text: string(cat), CallbackData: "rights_cat:" + string(cat)},
		})
	}
	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func FormatCommandsByCategory(category cmd.Category, commands []*cmd.Command, perms map[string]model.Status) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Категория: %s\n\n", category))

	for _, c := range commands {
		if c.GetCategory() != category {
			continue
		}

		status := c.GetDefaultStatus()
		if s, ok := perms[c.GetKey()]; ok {
			status = s
		}

		aliases := c.GetAliases()
		formattedAliases := make([]string, len(aliases))
		for i, a := range aliases {
			formattedAliases[i] = "<code>" + a + "</code>"
		}

		sb.WriteString(fmt.Sprintf("%s %s (%s)\n",
			helpers.StatusEmoji(status),
			c.GetDescription(),
			strings.Join(formattedAliases, ", "),
		))
	}

	sb.WriteString("\nВыберите команду для изменения прав:")
	return sb.String()
}

func GetCommandsByCategoryKeyboard(category cmd.Category, commands []*cmd.Command, perms map[string]model.Status) gotgbot.InlineKeyboardMarkup {
	var rows [][]gotgbot.InlineKeyboardButton
	for _, c := range commands {
		if c.GetCategory() != category {
			continue
		}

		status := c.GetDefaultStatus()
		if s, ok := perms[c.GetKey()]; ok {
			status = s
		}

		rows = append(rows, []gotgbot.InlineKeyboardButton{
			{Text: c.GetDescription(), CallbackData: "rights_edit:" + c.GetKey(), IconCustomEmojiId: helpers.StatusEmojiId(status)},
		})
	}

	rows = append(rows, []gotgbot.InlineKeyboardButton{
		{Text: "К категориям", CallbackData: "rights_list", Style: "danger"},
	})

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func FormatEditCommandRights(key string, currentStatus model.Status, commands []*cmd.Command) string {
	var config *cmd.Command
	for _, c := range commands {
		if c.GetKey() == key {
			config = c
			break
		}
	}

	if config == nil {
		return "Ошибка: команда не найдена"
	}

	aliases := config.GetAliases()
	formattedAliases := make([]string, len(aliases))
	for i, a := range aliases {
		formattedAliases[i] = "<code>" + a + "</code>"
	}

	return fmt.Sprintf("⚙️ %s\nАлиасы: %s\n\nТекущий уровень: %s %s\n\nВыберите новый уровень доступа:",
		config.GetDescription(),
		strings.Join(formattedAliases, ", "),
		helpers.StatusEmoji(currentStatus),
		currentStatus.String(),
	)
}

func GetEditRightsKeyboard(key string, currentStatus model.Status, commands []*cmd.Command) gotgbot.InlineKeyboardMarkup {
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
			IconCustomEmojiId: helpers.StatusEmojiId(s),
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
		if c.GetKey() == key {
			backCategory = string(c.GetCategory())
			break
		}
	}

	rows = append(rows, []gotgbot.InlineKeyboardButton{
		{Text: "Назад", CallbackData: "rights_cat:" + backCategory, Style: "danger"},
	})

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: rows}
}
