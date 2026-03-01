package helpers

import (
	"fmt"
	"strings"
	"time"
)

var MoscowLocation *time.Location

func init() {
	var err error
	MoscowLocation, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		MoscowLocation = time.FixedZone("MSK", 3*3600)
	}
}

func FormatToHumanDateTime(date time.Time) string {
	date = date.In(MoscowLocation)
	now := time.Now().In(MoscowLocation)

	months := [...]string{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}

	hour := date.Hour()
	minute := date.Minute()

	timePart := fmt.Sprintf("%02d:%02d", hour, minute)

	if now.Year() == date.Year() && now.YearDay() == date.YearDay() {
		return fmt.Sprintf("сегодня в %s", timePart)
	}

	if now.Year() == date.Year() {
		return fmt.Sprintf("%d %s, %s", date.Day(), months[date.Month()-1], timePart)
	}

	return fmt.Sprintf("%d %s %d, %s", date.Day(), months[date.Month()-1], date.Year(), timePart)
}

func FormatToHumanDate(date time.Time) string {
	date = date.In(MoscowLocation)
	now := time.Now().In(MoscowLocation)

	months := [...]string{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}

	if now.Year() == date.Year() && now.YearDay() == date.YearDay() {
		return fmt.Sprintf("сегодня")
	}

	if now.Year() == date.Year() {
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
