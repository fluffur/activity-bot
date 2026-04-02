package command

import (
	"activity-bot/internal/model"
)

type Category string

const (
	CategoryGeneral    Category = "Общие"
	CategoryProfile    Category = "Профиль"
	CategoryModeration Category = "Модерация"
	CategoryCall       Category = "Каллы"
	CategorySettings   Category = "Настройки"
	CategoryStats      Category = "Статистика"
	CategoryAdmin      Category = "Администрирование"
)

func GetDefaultStatus(commands []*Command, key string) model.Status {
	for _, cmd := range commands {
		if cmd.Name() == key {
			return cmd.RequiredStatus()
		}
	}
	return model.StatusAdmin
}
