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
