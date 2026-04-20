package helpers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gotd/td/telegram/message/entity"
)

var MoscowLocation *time.Location

func init() {
	var err error
	MoscowLocation, err = time.LoadLocation("Europe/Moscow")
	if err != nil {
		MoscowLocation = time.FixedZone("MSK", 3*3600)
	}
	time.Local = MoscowLocation
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
func formatLastSeenHuman(t time.Time) string {
	t = t.Local()
	now := time.Now()

	d := now.Sub(t)

	if d < time.Hour {
		minutes := int(d.Minutes())
		if minutes <= 1 {
			return "только что"
		}
		return fmt.Sprintf("%d %s",
			minutes,
			pluralRu(minutes, "минуту", "минуты", "минут"),
		)
	}

	totalHours := int(d.Hours())

	years := totalHours / (24 * 365)
	months := (totalHours % (24 * 365)) / (24 * 30)
	days := (totalHours % (24 * 30)) / 24
	hours := totalHours % 24

	var parts []string

	if years > 0 {
		parts = append(parts,
			fmt.Sprintf("%d %s", years, pluralRu(years, "год", "года", "лет")),
		)
	}

	if months > 0 {
		parts = append(parts,
			fmt.Sprintf("%d %s", months, pluralRu(months, "месяц", "месяца", "месяцев")),
		)
	}

	if days > 0 && years == 0 {
		parts = append(parts,
			fmt.Sprintf("%d %s", days, pluralRu(days, "день", "дня", "дней")),
		)
	}

	if hours > 0 && years == 0 && months == 0 {
		parts = append(parts,
			fmt.Sprintf("%d %s", hours, pluralRu(hours, "час", "часа", "часов")),
		)
	}

	return strings.Join(parts, " ")
}

func FormatLastSeenPlain(t time.Time) string {
	return formatLastSeenHuman(t)
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

func FormattedDate(eb *entity.Builder, date time.Time) {
	if date.Year() > time.Now().Year()+8 {
		eb.Plain(FormatToHumanDateTime(date))
		return
	}
	eb.FormattedDate(FormatToHumanDateTime(date),
		false,
		true,
		false,
		true,
		false,
		true,
		int(date.Unix()),
	)
}

func FormatToHumanDateTime(date time.Time) string {
	date = date.Local()
	now := time.Now()

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

	return text
}
