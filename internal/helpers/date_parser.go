package helpers

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

var monthNames = map[string]time.Month{
	"января":   time.January,
	"февраля":  time.February,
	"марта":    time.March,
	"апреля":   time.April,
	"мая":      time.May,
	"июня":     time.June,
	"июля":     time.July,
	"августа":  time.August,
	"сентября": time.September,
	"октября":  time.October,
	"ноября":   time.November,
	"декабря":  time.December,
}

type DateParser struct {
	now func() time.Time
}

func NewDateParser() *DateParser {
	return &DateParser{now: time.Now}
}

func (p *DateParser) Parse(arg string) (time.Time, bool) {
	arg = strings.TrimSpace(strings.ToLower(arg))

	switch arg {
	case "сегодня":
		return p.startOfDay(p.now()), true
	case "завтра":
		return p.startOfDay(p.now().AddDate(0, 0, 1)), true
	}

	if t, ok := p.parseRussianDate(arg); ok {
		return t, true
	}

	if t, ok := p.parseStandardDate(arg); ok {
		return t, true
	}

	if days, ok := p.ParseDays(arg); ok {
		return p.startOfDay(p.now()).AddDate(0, 0, days), true
	}

	return time.Time{}, false
}

func (p *DateParser) parseRussianDate(arg string) (time.Time, bool) {
	re := regexp.MustCompile(`^(\d{1,2})\s+([а-я]+)(?:\s+(\d{2,4})(?:г(?:ода)?)?)?$`)
	m := re.FindStringSubmatch(arg)
	if len(m) == 0 {
		return time.Time{}, false
	}

	day, _ := strconv.Atoi(m[1])
	monthStr := m[2]
	month, ok := monthNames[monthStr]
	if !ok {
		return time.Time{}, false
	}

	year := p.now().Year()
	if m[3] != "" {
		y, err := strconv.Atoi(m[3])
		if err != nil {
			return time.Time{}, false
		}
		year = y
	}

	return time.Date(year, month, day, 0, 0, 0, 0, p.now().Location()), true
}

func (p *DateParser) parseStandardDate(arg string) (time.Time, bool) {
	layouts := []string{"02.01.2006", "02.01", "2006-01-02"}
	now := p.now()

	for _, layout := range layouts {
		t, err := time.Parse(layout, arg)
		if err != nil {
			continue
		}
		if layout == "02.01" {
			t = time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
		}
		return t, true
	}
	return time.Time{}, false
}

func (p *DateParser) ParseDays(arg string) (int, bool) {
	arg = strings.TrimSpace(strings.ToLower(arg))
	if days, err := strconv.Atoi(arg); err == nil {
		return days, true
	}

	re := regexp.MustCompile(`^(?:(\d+)\s*)?(день|дня|дней|неделя|недели|недель|месяц|месяца|месяцев)$`)
	m := re.FindStringSubmatch(arg)
	if len(m) == 0 {
		return 0, false
	}

	count := 1
	if m[1] != "" {
		count, _ = strconv.Atoi(m[1])
	}

	switch m[2] {
	case "день", "дня", "дней":
		return count, true
	case "неделя", "недели", "недель":
		return count * 7, true
	case "месяц", "месяца", "месяцев":
		return count * 30, true
	}

	return 0, false
}

func (p *DateParser) startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
