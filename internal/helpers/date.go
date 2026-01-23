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
