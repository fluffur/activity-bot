package view

import (
	"activity-bot/internal/helpers"
	"fmt"
)

func FormatNorm(norm int, normBan int) string {
	banInfo := ""
	if normBan != 0 {
		banInfo = fmt.Sprintf("\nЕсли сообщений вместе с тем меньше, чем %d, то выдаётся бан", normBan)
	}

	return fmt.Sprintf("Если сообщений меньше %d, то выдается варн. %s", norm, banInfo)
}

func FormatNormSet(norm int, action string) string {
	return fmt.Sprintf("Установлена новая норма чата: %d на %s", norm, action)
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
