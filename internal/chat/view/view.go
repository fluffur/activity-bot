package view

import (
	"activity-bot/internal/helpers"
	"fmt"
	"time"
)

func FormatNorm(norm, normBan int32) string {
	if norm == 0 && normBan == 0 {
		return fmt.Sprintf("Норма еще не установлена\n\nПопробуйте <code>!норма 100 варн</code> или <code>!норма 40 бан</code>")
	}
	if norm == normBan {
		return fmt.Sprintf("Норма: %d сообщений", norm)
	}
	banInfo := ""
	if normBan != 0 {
		banInfo = fmt.Sprintf(", меньше %d бан", normBan)
	}

	return fmt.Sprintf("Норма меньше %d варн%s", norm, banInfo)
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

func FormatWeekStart(day int, time string) string {
	return FormatWeek("Начало недели", day, time)

}

func FormatWeekStartSet(day int, time string) string {
	return FormatWeek("Начало недели изменено", day, time)
}

func FormatWeek(msg string, day int, timeStr string) string {
	now := time.Now()

	var h, m int
	fmt.Sscanf(timeStr, "%d:%d", &h, &m)

	target := now
	for int(target.Weekday()) != day%7 {
		target = target.AddDate(0, 0, 1)
	}

	target = time.Date(
		target.Year(),
		target.Month(),
		target.Day(),
		h,
		m,
		0,
		0,
		time.Local,
	)

	return fmt.Sprintf(
		"📅 %s: <tg-time unix=\"%d\">%s в %s</tg-time>",
		msg,
		target.Unix(),
		helpers.FormatWeekStartDay(day),
		timeStr,
	)
}
