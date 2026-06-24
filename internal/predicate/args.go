package predicate

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/middleware/cctx"
	"context"
	"strconv"
	"strings"
	"time"

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

type ArgChecker struct {
	chatMemberRepository chatmember.Repository
}

func NewArgChecker(cmr chatmember.Repository) *ArgChecker {
	return &ArgChecker{cmr}
}

func (a *ArgChecker) WithArgs(rules ...Arg) botapi.Predicate {
	return func(c *botapi.Context) bool {
		argsMessage := Args(c)

		if argsMessage == nil && !allRulesAreOptional(rules...) && len(rules) != 0 {
			return false
		}

		text, entities := argsMessage.TextAndEntities()
		if strings.TrimSpace(text) == "" && !allRulesAreOptional(rules...) && len(rules) != 0 {
			return false
		}

		ch, err := cctx.Chat(c.Context)
		if err != nil {
			log.For(c.Bot.Logger()).Error(c.Context, "with args: chat ctx missing", log.Error(err))
			return false
		}

		var usedOffsets []Offset
		parsed := ParsedArgs{}

		usedOffsets = resolveUserEntities(c.Context, a.chatMemberRepository, ch.ID, text, entities, &parsed, usedOffsets)

		for _, rule := range rules {
			count := rule.Count
			if count == ArgCountVariadic {
				count = 50
			}

			parsedCount := 0

			if rule.Type == ArgTypeText {
				toks := getFreeTokens(text, usedOffsets)
				log.For(c.Bot.Logger()).Debug(c.Context, "found free tokens", log.Int("count", count))
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
								if cm, err := a.chatMemberRepository.Get(c.Context, ch.ID, id); err == nil {
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

func allRulesAreOptional(rules ...Arg) bool {
	for _, rule := range rules {
		if !rule.Optional {
			return false
		}
	}
	return true
}
