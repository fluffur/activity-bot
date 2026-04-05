package view

import (
	"activity-bot/internal/helpers"
	"fmt"
	"time"

	"github.com/gotd/td/telegram/message/entity"
)

func WriteNorm(eb *entity.Builder, norm, normBan int32) {
	if norm == 0 && normBan == 0 {
		eb.Plain("Норма еще не установлена\n\nПопробуйте ")
		eb.Code("!норма 100 варн")
		eb.Plain(" или ")
		eb.Code("!норма 40 бан")
		return
	}
	if norm == normBan {
		eb.Plain(fmt.Sprintf("Норма: %d сообщений", norm))
		return
	}
	banInfo := ""
	if normBan != 0 {
		banInfo = fmt.Sprintf(", меньше %d бан", normBan)
	}

	eb.Plain(fmt.Sprintf("Норма меньше %d варн%s", norm, banInfo))
}

func FormatNormSet(norm int, action string) string {
	return fmt.Sprintf("Установлена новая норма чата: %d на %s", norm, action)
}

func FormatNewbieThreshold(days int32) string {
	if days == 0 {
		return fmt.Sprintf("Срок для новичков ещё не указан")
	}
	return fmt.Sprintf("Участники считаются новичками первые %d %s", days, helpers.PluralizeDays(int(days)))
}

func FormatNewbieThresholdSet(days int32) string {
	return fmt.Sprintf("Теперь участники считаются новичками первые %d %s", days, helpers.PluralizeDays(int(days)))
}

func FormatPrompt(prompt string) string {
	return fmt.Sprintf("Промпт: \"%s\"", prompt)
}

func WritePrefix(eb *entity.Builder, prefix string) {
	if prefix == "" {
		eb.Plain("В этом чате не установлен кастомный префикс.")
		return
	}
	eb.Plain("Текущий префикс чата: ")
	eb.Code(prefix)
}

func WritePrefixSet(eb *entity.Builder, prefix string) {
	eb.Plain("Установлен новый префикс чата: ")
	eb.Code(prefix)
}

func WritePrefixlessToggle(eb *entity.Builder, enabled bool) {
	eb.Plain("Теперь ")
	WritePrefixlessStatus(eb, enabled)
}

func WritePrefixlessStatus(eb *entity.Builder, enabled bool) {
	if !enabled {
		eb.Plain("Бот не отвечает на сообщения без префиксов")
		return
	}
	eb.Plain("Бот отвечает на сообщения без префиксов")
}

func WriteWeekStart(eb *entity.Builder, day int, timeStr string) {
	WriteWeek(eb, "Начало недели", day, timeStr)
}

func WriteWeekStartSet(eb *entity.Builder, day int, timeStr string) {
	WriteWeek(eb, "Начало недели изменено", day, timeStr)
}

func WriteWeek(eb *entity.Builder, msg string, day int, timeStr string) {
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

	eb.Plain("📅 " + msg + ": ")
	helpers.FormattedDate(eb, target)
}
