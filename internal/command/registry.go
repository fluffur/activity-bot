package command

import (
	"activity-bot/internal/model"
)

type Category string

const (
	CategoryModeration Category = "Модерация"
	CategoryCall       Category = "Рассылки"
	CategorySettings   Category = "Настройки"
	CategoryStats      Category = "Статистика"
)

func GetDefaultStatus(commands []*Command, key string) model.Status {
	for _, cmd := range commands {
		if cmd.Name() == key {
			return cmd.RequiredStatus()
		}
	}
	return model.StatusAdmin
}
