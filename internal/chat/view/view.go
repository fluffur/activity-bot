package view

import (
	"activity-bot/internal/helpers"
	"fmt"
)

func FormatNorm(norm int) string {
	return fmt.Sprintf("Норма чата: %d сообщений", norm)
}

func FormatNormSet(norm int) string {
	return fmt.Sprintf("Установлена новая норма чата: %d", norm)
}

func FormatNewbieThreshold(days int) string {
	return fmt.Sprintf("Пользователи считаются новичками первые %d %s", days, helpers.PluralizeDays(days))
}

func FormatNewbieThresholdSet(days int) string {
	return fmt.Sprintf("Теперь пользователи считаются новичками первые %d %s", days, helpers.PluralizeDays(days))
}

func FormatPrompt(prompt string) string {
	return fmt.Sprintf("Промпт: \"%s\"", prompt)
}
