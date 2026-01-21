package chat

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

type DateParser struct {
	now func() time.Time
}

func NewDateParser() *DateParser {
	return &DateParser{
		now: time.Now,
	}
}

func (p *DateParser) Parse(arg string) (time.Time, bool) {
	arg = strings.TrimSpace(strings.ToLower(arg))

	switch arg {
	case "сегодня":
		return p.startOfDay(p.now()), true
	case "завтра":
		return p.startOfDay(p.now().AddDate(0, 0, 1)), true
	}

	if t, ok := p.parseDate(arg); ok {
		return t, true
	}

	if t, ok := p.parsePeriod(arg); ok {
		return t, true
	}

	return time.Time{}, false
}

func (p *DateParser) parseDate(arg string) (time.Time, bool) {
	layouts := []string{
		"02.01.2006",
		"02.01",
		"2006-01-02",
	}

	now := p.now()

	for _, layout := range layouts {
		t, err := time.Parse(layout, arg)
		if err != nil {
			continue
		}

		if layout == "02.01" {
			t = time.Date(
				now.Year(),
				t.Month(),
				t.Day(),
				0, 0, 0, 0,
				now.Location(),
			)
		}

		return t, true
	}

	return time.Time{}, false
}

func (p *DateParser) parsePeriod(arg string) (time.Time, bool) {
	re := regexp.MustCompile(`^(?:(\d+)\s*)?(день|дня|дней|неделя|недели|недель|месяц|месяца|месяцев)$`)
	m := re.FindStringSubmatch(arg)
	if len(m) == 0 {
		return time.Time{}, false
	}

	count := 1
	if m[1] != "" {
		count, _ = strconv.Atoi(m[1])
	}

	now := p.startOfDay(p.now())

	switch m[2] {
	case "день", "дня", "дней":
		return now.AddDate(0, 0, count), true
	case "неделя", "недели", "недель":
		return now.AddDate(0, 0, count*7), true
	case "месяц", "месяца", "месяцев":
		return now.AddDate(0, count, 0), true
	}

	return time.Time{}, false
}

func (p *DateParser) startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
