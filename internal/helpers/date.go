package helpers

import (
	"fmt"
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

	switch {
	case d < time.Minute:
		return "только что"

	case d < time.Hour:
		minValue := int(d.Minutes())
		return fmt.Sprintf(
			"%d %s назад",
			minValue,
			pluralRu(minValue, "минута", "минуты", "минут"),
		)

	case d < 24*time.Hour:
		h := int(d.Hours())
		return fmt.Sprintf(
			"%d %s назад",
			h,
			pluralRu(h, "час", "часа", "часов"),
		)

	default:
		days := int(d.Hours() / 24)
		return fmt.Sprintf(
			"%d %s назад",
			days,
			pluralRu(days, "день", "дня", "дней"),
		)
	}
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
