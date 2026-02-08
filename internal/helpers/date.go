package helpers

import (
	"fmt"
	"strings"
	"time"
)

func FormatToHumanDate(date time.Time) string {
	months := [...]string{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}
	if time.Now().Year() == date.Year() {
		return fmt.Sprintf("%d %s", date.Day(), months[date.Month()-1])
	}

	return fmt.Sprintf("%d %s %d", date.Day(), months[date.Month()-1], date.Year())
}

func PluralizeDays(n int) string {
	nAbs := n % 100
	if nAbs >= 11 && nAbs <= 14 {
		return "дней"
	}

	switch n % 10 {
	case 1:
		return "день"
	case 2, 3, 4:
		return "дня"
	default:
		return "дней"
	}
}

func FormatLastSeen(t time.Time) string {
	d := time.Since(t)

	if d < time.Minute {
		return "только что"
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	var parts []string

	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", days, pluralRu(days, "день", "дня", "дней")))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", hours, pluralRu(hours, "час", "часа", "часов")))
	}
	if days == 0 && hours == 0 && minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d %s", minutes, pluralRu(minutes, "минута", "минуты", "минут")))
	}

	return strings.Join(parts, " ")
}

func pluralRu(n int, one, few, many string) string {
	n = n % 100
	if n >= 11 && n <= 14 {
		return many
	}

	switch n % 10 {
	case 1:
		return one
	case 2, 3, 4:
		return few
	default:
		return many
	}
}
