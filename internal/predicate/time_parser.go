package predicate

import (
	"strconv"
	"strings"
	"time"
)

func tryParseAdvancedDuration(toks []token) (time.Duration, int, bool) {
	if len(toks) == 0 {
		return 0, 0, false
	}

	t0 := strings.ToLower(toks[0].text)

	switch t0 {
	case "неделя", "неделю":
		return 7 * 24 * time.Hour, 1, true
	case "месяц":
		return 30 * 24 * time.Hour, 1, true
	case "год":
		return 365 * 24 * time.Hour, 1, true
	}

	if d, err := time.ParseDuration(t0); err == nil {
		return d, 1, true
	}

	if len(toks) >= 2 {
		val, err := strconv.ParseInt(t0, 10, 64)
		if err == nil {
			t1 := strings.ToLower(toks[1].text)

			var multiplier time.Duration

			switch {
			case strings.HasPrefix(t1, "мин"): // минут, минута, минуты
				multiplier = time.Minute
			case strings.HasPrefix(t1, "час"): // час, часа, часов
				multiplier = time.Hour
			case strings.HasPrefix(t1, "ден") || strings.HasPrefix(t1, "дн"): // день, дня, дней
				multiplier = 24 * time.Hour
			case strings.HasPrefix(t1, "недел"): // неделя, недели, недель
				multiplier = 7 * 24 * time.Hour
			case strings.HasPrefix(t1, "месяц"): // месяц, месяца, месяцев
				multiplier = 30 * 24 * time.Hour
			case strings.HasPrefix(t1, "год") || strings.HasPrefix(t1, "лет"): // год, года, лет
				multiplier = 365 * 24 * time.Hour
			}

			if multiplier > 0 {
				return time.Duration(val) * multiplier, 2, true
			}
		}
	}

	return 0, 0, false
}

func tryParseAdvancedDateTime(toks []token) (time.Time, int, bool) {
	if len(toks) == 0 {
		return time.Time{}, 0, false
	}

	t0 := toks[0].text
	loc, _ := time.LoadLocation("Europe/Moscow")
	now := time.Now().In(loc)

	// DD.MM.YYYY
	if t, err := time.ParseInLocation("02.01.2006", t0, loc); err == nil {
		return t, 1, true
	}

	// DD.MM
	if t, err := time.ParseInLocation("02.01", t0, loc); err == nil {
		parsedDate := time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
		return parsedDate, 1, true
	}

	// YYYY-MM-DD
	if t, err := time.ParseInLocation("2006-01-02", t0, loc); err == nil {
		return t, 1, true
	}

	// RFC3339
	if t, err := time.ParseInLocation(time.RFC3339, t0, loc); err == nil {
		return t, 1, true
	}

	if len(toks) >= 2 {
		day, err := strconv.Atoi(t0)
		if err == nil && day >= 1 && day <= 31 {
			monthStr := strings.ToLower(toks[1].text)
			month := matchRussianMonth(monthStr)

			if month != 0 {
				year := now.Year()
				consumed := 2

				if len(toks) >= 3 {
					if y, err := strconv.Atoi(toks[2].text); err == nil && y > 2000 && y < 2100 {
						year = y
						consumed = 3
					}
				}

				parsedDate := time.Date(year, month, day, 0, 0, 0, 0, now.Location())

				return parsedDate, consumed, true
			}
		}
	}

	return time.Time{}, 0, false
}

func matchRussianMonth(m string) time.Month {
	switch {
	case strings.HasPrefix(m, "янв"):
		return time.January
	case strings.HasPrefix(m, "фев"):
		return time.February
	case strings.HasPrefix(m, "мар"):
		return time.March
	case strings.HasPrefix(m, "апр"):
		return time.April
	case strings.HasPrefix(m, "май") || strings.HasPrefix(m, "мае"):
		return time.May
	case strings.HasPrefix(m, "июн"):
		return time.June
	case strings.HasPrefix(m, "июл"):
		return time.July
	case strings.HasPrefix(m, "авг"):
		return time.August
	case strings.HasPrefix(m, "сен"):
		return time.September
	case strings.HasPrefix(m, "окт"):
		return time.October
	case strings.HasPrefix(m, "ноя"):
		return time.November
	case strings.HasPrefix(m, "дек"):
		return time.December
	}

	return 0
}
