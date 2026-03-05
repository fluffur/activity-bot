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

	var text string
	if now.Year() == date.Year() && now.YearDay() == date.YearDay() {
		text = "сегодня"
	} else if now.Year() == date.Year() {
		text = fmt.Sprintf("%d %s", date.Day(), months[date.Month()-1])
	} else {
		text = fmt.Sprintf("%d %s %d", date.Day(), months[date.Month()-1], date.Year())
	}

	return fmt.Sprintf("<tg-time unix=\"%d\">%s</tg-time>", date.Unix(), text)
}

func FormatToHumanDate(date time.Time) string {
	date = date.In(MoscowLocation)
	now := time.Now().In(MoscowLocation)

	months := [...]string{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}

	var text string
	if now.Year() == date.Year() && now.YearDay() == date.YearDay() {
		text = "сегодня"
	} else if now.Year() == date.Year() {
		text = fmt.Sprintf("%d %s", date.Day(), months[date.Month()-1])
	} else {
		text = fmt.Sprintf("%d %s %d", date.Day(), months[date.Month()-1], date.Year())
	}

	return fmt.Sprintf("<tg-time unix=\"%d\">%s</tg-time>", date.Unix(), text)
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
	t = t.In(MoscowLocation)
	now := time.Now().In(MoscowLocation)

	d := now.Sub(t)

	var human string

	if d < time.Hour {
		human = fmt.Sprintf("<tg-time unix=\"%d\">меньше часа назад</tg-time>)", t.Unix())
	} else {
		days := int(d.Hours()) / 24
		hours := int(d.Hours()) % 24

		var parts []string

		if days > 0 {
			parts = append(parts,
				fmt.Sprintf("%d %s", days, pluralRu(days, "день", "дня", "дней")),
			)
		}
		if hours > 0 {
			parts = append(parts,
				fmt.Sprintf("%d %s", hours, pluralRu(hours, "час", "часа", "часов")),
			)
		}

		human = strings.Join(parts, " ") + " назад"
	}

	return fmt.Sprintf(
		"<tg-time unix=\"%d\">%s</tg-time>",
		t.Unix(),
		human,
	)
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

func FormatWeekStartDay(day int) string {
	days := [...]string{
		"понедельник", "вторник", "среда", "четверг", "пятница", "суббота", "воскресенье",
	}
	if day < 1 || day > 7 {
		return "неизвестно"
	}
	return days[day-1]
}

func TimeToMicroseconds(s string) int64 {
	var h, m int
	fmt.Sscanf(s, "%d:%d", &h, &m)
	return int64(h)*3600*1000000 + int64(m)*60*1000000
}

func MicrosecondsToTime(micros int64) string {
	seconds := micros / 1000000
	h := seconds / 3600
	m := (seconds % 3600) / 60
	return fmt.Sprintf("%02d:%02d", h, m)
}
