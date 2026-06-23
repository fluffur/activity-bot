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
					for _, tok := range toks {
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

						if rule.Type == ArgTypeDateTime {
							if t, err := parseFlexDateTime(tok.text); err == nil {
								parsed.DateTimes = append(parsed.DateTimes, t)
								usedOffsets = append(usedOffsets, Offset{tok.start, tok.end})
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

func parseFlexDateTime(tok string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, tok); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02", tok); err == nil {
		return t, nil
	}
	return time.Time{}, strconv.ErrSyntax
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
