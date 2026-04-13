package view

import (
	"activity-bot/internal/command"
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"strconv"
	"strings"

	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
)

func WriteCategories(eb *entity.Builder) {
	eb.Plain("⚖️ ")
	eb.Bold("Настройка прав доступа команд")
	eb.Plain("\n\nВыберите категорию команд для настройки:")
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

func WriteCommandsByCategory(eb *entity.Builder, category command.Category, commands []*command.Command, perms map[string]model.Status) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Категория: %s\n\n", category))
	eb.Plain(sb.String())
	sb.Reset()

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
			formattedAliases[i] = a
		}

		helpers.WriteStatusEmoji(eb, status)
		eb.Plain(" ")
		eb.Plain(c.Description())
		eb.Plain(" (")
		for i, alias := range formattedAliases {
			if i > 0 {
				eb.Plain(", ")
			}
			eb.Code(alias)
		}
		eb.Plain(")\n")
	}

	eb.Plain("\nВыберите команду для изменения прав:")
}

func GetCommandsByCategoryKeyboard(category command.Category, commands []*command.Command, perms map[string]model.Status) tg.ReplyMarkupClass {
	var rows []tg.KeyboardButtonRow
	for _, c := range commands {
		if c.Category() != category {
			continue
		}

		icon, _ := strconv.ParseInt(helpers.StatusEmojiID(c.RequiredStatus()), 10, 64)
		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonCallback{
					Text: c.Description(),
					Data: []byte("rights_edit:" + c.Name()),
					Style: tg.KeyboardButtonStyle{
						Icon: icon,
					},
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
	eb := &entity.Builder{}
	WriteEditCommandRights(eb, key, currentStatus, commands)
	res, _ := eb.Complete()
	return res
}

func WriteEditCommandRights(eb *entity.Builder, key string, currentStatus model.Status, commands []*command.Command) {
	var config *command.Command
	for _, c := range commands {
		if c.Name() == key {
			config = c
			break
		}
	}

	if config == nil {
		eb.Plain("Ошибка: команда не найдена")
		return
	}

	aliases := append(config.Aliases(), config.Name())
	eb.Plain("⚙️ ")
	eb.Plain(config.Description())
	eb.Plain("\nСинонимы: ")
	for i, a := range aliases {
		if i > 0 {
			eb.Plain(", ")
		}
		eb.Code(a)
	}
	eb.Plain("\n\nТекущий уровень: ")
	helpers.WriteStatusEmoji(eb, currentStatus)
	eb.Plain(" ")
	eb.Plain(currentStatus.String())
	eb.Plain("\n\nВыберите новый уровень доступа:")
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
