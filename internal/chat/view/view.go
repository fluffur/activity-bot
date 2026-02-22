package view

import (
	"activity-bot/internal/helpers"
	"fmt"
)

func FormatNorm(norm, normBan int32) string {
	if norm == normBan {
		return fmt.Sprintf("Норма: %d сообщений", norm)
	}
	banInfo := ""
	if normBan != 0 {
		banInfo = fmt.Sprintf("\nЕсли сообщений меньше, чем %d, то выдаётся бан", normBan)
	}

	return fmt.Sprintf("Если сообщений меньше %d, то выдается варн. %s", norm, banInfo)
}

func FormatNormSet(norm int, action string) string {
	return fmt.Sprintf("Установлена новая норма чата: %d на %s", norm, action)
}

func FormatNewbieThreshold(days int) string {
	return fmt.Sprintf("Участники считаются новичками первые %d %s", days, helpers.PluralizeDays(days))
}

func FormatNewbieThresholdSet(days int) string {
	return fmt.Sprintf("Теперь участники считаются новичками первые %d %s", days, helpers.PluralizeDays(days))
}

func FormatPrompt(prompt string) string {
	return fmt.Sprintf("Промпт: \"%s\"", prompt)
}

func FormatPrefix(prefix string) string {
	if prefix == "" {
		return "В этом чате не установлен кастомный префикс."
	}
	return fmt.Sprintf("Текущий префикс чата: `%s`", prefix)
}

func FormatPrefixSet(prefix string) string {
	return fmt.Sprintf("Установлен новый префикс чата: `%s`", prefix)
}

func FormatPrefixlessToggle(enabled bool) string {
	return "Теперь " + FormatPrefixlessStatus(enabled)
}

func FormatPrefixlessStatus(enabled bool) string {
	if !enabled {
		return "Бот не отвечает на сообщения без префиксов"
	}
	return "Бот отвечает на сообщения без префиксов"
}
