package predicate

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/middleware/cctx"
	"context"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/gotd/log"

	"github.com/gotd/botapi"
)

func NoArgs() botapi.Predicate {
	return func(c *botapi.Context) bool {
		args := Args(c)
		return strings.TrimSpace(args.Text) == ""
	}
}

type ArgType string

const (
	ArgTypeUser     ArgType = "user"
	ArgTypeNumber   ArgType = "number"
	ArgTypeDuration ArgType = "duration"
	ArgTypeDateTime ArgType = "datetime"
	ArgTypeText     ArgType = "text"
)

const ArgCountVariadic = -1

type Arg struct {
	Type      ArgType
	Optional  bool
	Count     int
	OnNextRow bool
}

type ParsedArgs struct {
	Users     []chatmember.ChatMember
	Numbers   []int64
	Durations []time.Duration
	DateTimes []time.Time
	Texts     []string
}

type parsedArgsKey struct{}

var commandParsedArgsKey = parsedArgsKey{}

func GetParsedArgs(c *botapi.Context) (ParsedArgs, bool) {
	val, ok := c.Context.Value(commandParsedArgsKey).(ParsedArgs)
	return val, ok
}

type Offset struct {
	Start int
	End   int
}

type token struct {
	text  string
	start int
	end   int
}

func WithArgs(repo chatmember.Repository, rules ...Arg) botapi.Predicate {
	return func(c *botapi.Context) bool {
		argsMessage := Args(c)
		log.For(c.Bot.Logger()).Info(c.Context, "handling args", log.Any("args", argsMessage))
		if argsMessage == nil {
			return len(rules) == 0
		}

		text, entities := argsMessage.TextAndEntities()
		if strings.TrimSpace(text) == "" {
			return len(rules) == 0
		}

		ch, err := cctx.Chat(c.Context)
		if err != nil {
			log.For(c.Bot.Logger()).Error(c.Context, "with args: chat ctx missing", log.Error(err))
			return false
		}

		var usedOffsets []Offset
		parsed := ParsedArgs{}

		usedOffsets = resolveUserEntities(c.Context, repo, ch.ID, text, entities, &parsed, usedOffsets)

		for _, rule := range rules {
			count := rule.Count
			if count == ArgCountVariadic {
				count = 50
			}

			parsedCount := 0

			if rule.Type == ArgTypeText {
				toks := getFreeTokens(text, usedOffsets)
				if len(toks) == 0 {
					if rule.Optional {
						continue
					}
					return false
				}

				var textParts []string
				if rule.OnNextRow {
					newLineIdx := strings.Index(text, "\n")
					for _, tok := range toks {
						if newLineIdx != -1 && tok.start > newLineIdx {
							textParts = append(textParts, tok.text)
							usedOffsets = append(usedOffsets, Offset{tok.start, tok.end})
						}
					}
				} else {
					for _, tok := range toks {
						textParts = append(textParts, tok.text)
						usedOffsets = append(usedOffsets, Offset{tok.start, tok.end})
					}
				}

				joinedText := strings.Join(textParts, " ")
				if joinedText != "" {
					parsed.Texts = append(parsed.Texts, joinedText)
					parsedCount++
				}
			} else {
				for i := 0; i < count; i++ {
					toks := getFreeTokens(text, usedOffsets)
					if len(toks) == 0 {
						break
					}

					matched := false
					for idx, tok := range toks {
						if rule.Type == ArgTypeUser {
							if id, err := strconv.ParseInt(tok.text, 10, 64); err == nil {
								if cm, err := repo.Get(c.Context, ch.ID, id); err == nil {
									parsed.Users = append(parsed.Users, cm)
									usedOffsets = append(usedOffsets, Offset{tok.start, tok.end})
									parsedCount++
									matched = true
									break
								}
							}
						}

						if rule.Type == ArgTypeDuration {
							if d, consumed, ok := tryParseAdvancedDuration(toks[idx:]); ok {
								parsed.Durations = append(parsed.Durations, d)
								for k := 0; k < consumed; k++ {
									usedOffsets = append(usedOffsets, Offset{toks[idx+k].start, toks[idx+k].end})
								}
								parsedCount++
								matched = true
								break
							}
						}

						if rule.Type == ArgTypeDateTime {
							if t, consumed, ok := tryParseAdvancedDateTime(toks[idx:]); ok {
								parsed.DateTimes = append(parsed.DateTimes, t)
								for k := 0; k < consumed; k++ {
									usedOffsets = append(usedOffsets, Offset{toks[idx+k].start, toks[idx+k].end})
								}
								parsedCount++
								matched = true
								break
							}
						}

						if rule.Type == ArgTypeNumber {
							if num, err := strconv.ParseInt(tok.text, 10, 64); err == nil {
								parsed.Numbers = append(parsed.Numbers, num)
								usedOffsets = append(usedOffsets, Offset{tok.start, tok.end})
								parsedCount++
								matched = true
								break
							}
						}
					}

					if !matched {
						break
					}
				}
			}

			if parsedCount == 0 && !rule.Optional {
				return false
			}
		}

		c.Context = context.WithValue(c.Context, commandParsedArgsKey, parsed)
		return true
	}
}

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

// tryParseAdvancedDateTime распознает "1 июля", "08.08", "07.07.2028"
func tryParseAdvancedDateTime(toks []token) (time.Time, int, bool) {
	if len(toks) == 0 {
		return time.Time{}, 0, false
	}

	t0 := toks[0].text
	now := time.Now()

	// Сначала проверяем стандартные форматы в рамках ОДНОГО токена
	// DD.MM.YYYY
	if t, err := time.Parse("02.01.2006", t0); err == nil {
		return t, 1, true
	}
	// DD.MM (подставляем текущий год)
	if t, err := time.Parse("02.01", t0); err == nil {
		parsedDate := time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location())
		return parsedDate, 1, true
	}
	// YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", t0); err == nil {
		return t, 1, true
	}
	// RFC3339
	if t, err := time.Parse(time.RFC3339, t0); err == nil {
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

func resolveUserEntities(
	ctx context.Context,
	repo chatmember.Repository,
	chatID int64,
	text string,
	entities []botapi.MessageEntity,
	parsed *ParsedArgs,
	used []Offset,
) []Offset {
	for _, entity := range entities {
		entityStr := entityTextUTF16(text, entity.Offset, entity.Length)
		if entityStr == "" {
			continue
		}

		u16 := utf16.Encode([]rune(text))
		byteStart := len(string(utf16.Decode(u16[:entity.Offset])))
		byteEnd := byteStart + len(string(utf16.Decode(u16[entity.Offset:entity.Offset+entity.Length])))

		switch entity.Type {
		case botapi.EntityMention:
			username := strings.TrimPrefix(entityStr, "@")
			if cm, err := repo.GetByUsername(ctx, chatID, username); err == nil {
				parsed.Users = append(parsed.Users, cm)
				used = append(used, Offset{byteStart, byteEnd})
			}

		case botapi.EntityTextMention:
			if entity.User != nil {
				if cm, err := repo.Get(ctx, chatID, entity.User.ID); err == nil {
					parsed.Users = append(parsed.Users, cm)
					used = append(used, Offset{byteStart, byteEnd})
				}
			}

		case botapi.EntityURL:
			if strings.Contains(entityStr, "t.me/") {
				parts := strings.Split(entityStr, "/")
				username := strings.TrimPrefix(parts[len(parts)-1], "@")
				if cm, err := repo.GetByUsername(ctx, chatID, username); err == nil {
					parsed.Users = append(parsed.Users, cm)
					used = append(used, Offset{byteStart, byteEnd})
				}
			}
		}
	}
	return used
}

func getFreeTokens(s string, used []Offset) []token {
	var tokens []token
	i := 0
	for i < len(s) {
		for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r') {
			i++
		}
		if i >= len(s) {
			break
		}
		j := i
		for j < len(s) && s[j] != ' ' && s[j] != '\t' && s[j] != '\n' && s[j] != '\r' {
			j++
		}

		overlapping := false
		for _, o := range used {
			if i < o.End && j > o.Start {
				overlapping = true
				break
			}
		}

		if !overlapping {
			tokens = append(tokens, token{text: s[i:j], start: i, end: j})
		}
		i = j
	}
	return tokens
}

func entityTextUTF16(text string, offset16, length16 int) string {
	u16 := utf16.Encode([]rune(text))
	if offset16 < 0 || offset16 >= len(u16) {
		return ""
	}
	end16 := offset16 + length16
	if end16 > len(u16) {
		end16 = len(u16)
	}
	return string(utf16.Decode(u16[offset16:end16]))
}
