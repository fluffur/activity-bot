package helpers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var weekdayMap = map[string]int{
	"пн": 1, "понедельник": 1,
	"вт": 2, "вторник": 2,
	"ср": 3, "среда": 3,
	"чт": 4, "четверг": 4,
	"пт": 5, "пятница": 5,
	"сб": 6, "суббота": 6,
	"вс": 7, "воскресенье": 7,
}

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
	return &DateParser{
		now: time.Now,
	}
}

func (p *DateParser) Parse(arg string) (time.Time, bool) {
	arg = strings.TrimSpace(strings.ToLower(arg))
	if t, ok := p.parseWeekday(arg); ok {
		return t, true
	}

	switch arg {
	case "сегодня":
		return p.now(), true
	case "завтра":
		return p.now().AddDate(0, 0, 1), true
	case "вчера":
		return p.now().AddDate(0, 0, -1), true
	}

	if t, ok := p.parseRussianDate(arg); ok {
		return t, true
	}

	if t, ok := p.parseStandardDate(arg); ok {
		return t, true
	}

	if duration, ok := p.ParseDuration(arg); ok {
		return p.now().Add(duration), true
	}

	if days, err := strconv.Atoi(arg); err == nil && days > 0 {
		return p.now().AddDate(0, 0, days), true
	}

	return time.Time{}, false
}

func (p *DateParser) parseWeekday(arg string) (time.Time, bool) {
	parts := strings.Fields(arg)
	if len(parts) == 0 {
		return time.Time{}, false
	}

	day, ok := weekdayMap[parts[0]]
	if !ok {
		return time.Time{}, false
	}

	now := p.now()

	var targetWD time.Weekday
	if day == 7 {
		targetWD = time.Sunday
	} else {
		targetWD = time.Weekday(day)
	}

	diff := int(targetWD - now.Weekday())
	if diff <= 0 {
		diff += 7
	}

	targetDate := now.AddDate(0, 0, diff)

	hour, minutes := 0, 0
	if len(parts) > 1 {
		if h, m, ok := parseTime(parts[1]); ok {
			hour = h
			minutes = m
		}
	}

	return time.Date(
		targetDate.Year(),
		targetDate.Month(),
		targetDate.Day(),
		hour,
		minutes,
		0,
		0,
		targetDate.Location(),
	), true
}

func parseTime(arg string) (int, int, bool) {
	var h, m int
	if _, err := fmt.Sscanf(arg, "%d:%d", &h, &m); err == nil {
		if h >= 0 && h <= 23 && m >= 0 && m <= 59 {
			return h, m, true
		}
	}
	if _, err := fmt.Sscanf(arg, "%d", &h); err == nil {
		if h >= 0 && h <= 23 {
			return h, 0, true
		}
	}
	return 0, 0, false
}

func (p *DateParser) ParseRange(args []string) (*time.Time, *time.Time, bool) {
	if len(args) == 0 {
		return nil, nil, false
	}

	fullStr := strings.TrimSpace(strings.ToLower(strings.Join(args, " ")))
	if fullStr == "" {
		return nil, nil, false
	}

	args = strings.Fields(fullStr)

	if len(args) == 1 {
		if days, err := strconv.Atoi(args[0]); err == nil && days > 0 {
			from := p.now().AddDate(0, 0, -days)

			return &from, nil, true
		}
	}

	if strings.Contains(fullStr, "-") {
		parts := strings.Split(fullStr, "-")
		if len(parts) == 2 {
			p1 := strings.TrimSpace(parts[0])
			p2 := strings.TrimSpace(parts[1])

			d1, err1 := strconv.Atoi(p1)
			d2, err2 := strconv.Atoi(p2)
			if err1 == nil && err2 == nil && d1 >= 1 && d1 <= 31 && d2 >= 1 && d2 <= 31 {
				now := p.now()

				startDay := d1
				endDay := d2
				if d1 > d2 {
					startDay = d2
					endDay = d1
				}
				from := time.Date(now.Year(), now.Month(), startDay, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location())
				to := time.Date(now.Year(), now.Month(), endDay, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location()).AddDate(0, 0, 1).Add(-time.Second)
				return &from, &to, true
			}

			from, ok1 := p.Parse(p1)
			to, ok2 := p.Parse(p2)
			if ok1 && ok2 {

				if from.After(to) {
					from, to = to, from
				}
				to = to.AddDate(0, 0, 1).Add(-time.Second)
				return &from, &to, true
			}
		}
	}

	var from, to *time.Time

	fromIdx := -1
	toIdx := -1
	for i, arg := range args {
		arg = strings.ToLower(arg)
		if (arg == "от" || arg == "с") && fromIdx == -1 {
			fromIdx = i
		} else if (arg == "до" || arg == "по") && toIdx == -1 {
			toIdx = i
		}
	}

	if fromIdx != -1 && toIdx != -1 {
		var fromPart, toPart string
		if fromIdx < toIdx {
			fromPart = strings.Join(args[fromIdx+1:toIdx], " ")
			toPart = strings.Join(args[toIdx+1:], " ")
		} else {
			toPart = strings.Join(args[toIdx+1:fromIdx], " ")
			fromPart = strings.Join(args[fromIdx+1:], " ")
		}

		if t, ok := p.Parse(fromPart); ok {
			from = &t
		}
		if t, ok := p.Parse(toPart); ok {
			t = t.AddDate(0, 0, 1).Add(-time.Second)
			to = &t
		}
	} else if fromIdx != -1 {
		fromPart := strings.Join(args[fromIdx+1:], " ")
		if t, ok := p.Parse(fromPart); ok {
			from = &t
		}
	} else if toIdx != -1 {
		toPart := strings.Join(args[toIdx+1:], " ")
		if t, ok := p.Parse(toPart); ok {
			t = t.AddDate(0, 0, 1).Add(-time.Second)
			to = &t
		}
	} else {
		if t, ok := p.Parse(fullStr); ok {
			from = &t
		}
	}

	if from != nil || to != nil {
		return from, to, true
	}

	return nil, nil, false
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

	now := p.now()
	year := now.Year()
	if m[3] != "" {
		y, err := strconv.Atoi(m[3])
		if err != nil {
			return time.Time{}, false
		}
		year = y
	}

	return time.Date(year, month, day, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location()), true
}

func (p *DateParser) parseStandardDate(arg string) (time.Time, bool) {
	layouts := []string{"02.01.2006", "02.01", "2006-01-02"}
	now := p.now()

	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, arg, now.Location())
		if err != nil {
			continue
		}
		year := t.Year()
		if layout == "02.01" {
			year = now.Year()
		}
		return time.Date(year, t.Month(), t.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location()), true
	}
	return time.Time{}, false
}
func (p *DateParser) ParseDuration(arg string) (time.Duration, bool) {
	arg = strings.TrimSpace(strings.ToLower(arg))

	re := regexp.MustCompile(`^(?:(\d+)\s*)?(день|дня|дней|неделя|недели|недель|месяц|месяца|месяцев|час|часа|часов|минута|минуты|минут|секунда|секунды|секунд)(?:\s+назад)?$`)
	m := re.FindStringSubmatch(arg)
	if len(m) == 0 {
		return 0, false
	}

	count := 1
	if m[1] != "" {
		count, _ = strconv.Atoi(m[1])
	}

	isAgo := strings.Contains(arg, "назад")
	if isAgo {
		count = -count
	}

	switch m[2] {
	case "день", "дня", "дней":
		return time.Duration(count) * 24 * time.Hour, true
	case "неделя", "недели", "недель":
		return time.Duration(count) * 7 * 24 * time.Hour, true
	case "месяц", "месяца", "месяцев":
		return time.Duration(count) * 30 * 24 * time.Hour, true
	case "час", "часа", "часов":
		return time.Duration(count) * time.Hour, true
	case "минута", "минуты", "минут":
		return time.Duration(count) * time.Minute, true
	case "секунда", "секунды", "секунд":
		return time.Duration(count) * time.Second, true
	}

	return 0, false
}
