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
	CategoryFun        Category = "Игровые"
)

func Categories() []Category {
	return []Category{
		CategoryGeneral,
		CategoryProfile,
		CategoryModeration,
		CategoryCall,
		CategorySettings,
		CategoryStats,
		CategoryAdmin,
		CategoryFun,
	}
}

func GetDefaultStatus(commands []*Command, key string) model.Status {
	for _, cmd := range commands {
		if cmd.Name() == key {
			return cmd.RequiredStatus()
		}
	}
	return model.StatusAdmin
}
