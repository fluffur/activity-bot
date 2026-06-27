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
		return args == nil || strings.TrimSpace(args.Text) == ""
	}
}

type RuleType string

const (
	RuleUser     RuleType = "user"
	RuleNumber   RuleType = "number"
	RuleDuration RuleType = "duration"
	RuleDateTime RuleType = "datetime"
	RuleText     RuleType = "text"
)

const RuleVariadic = -1

type Rule struct {
	Type         RuleType
	Optional     bool
	Count        int
	OnNextRow    bool
	TextValidate func(string) bool
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

type RuleChecker struct {
	chatMemberRepository chatmember.Repository
}

func NewRuleChecker(cmr chatmember.Repository) *RuleChecker {
	return &RuleChecker{cmr}
}

func (r *RuleChecker) With(rules ...Rule) botapi.Predicate {
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

		usedOffsets = resolveUserEntities(c.Context, r.chatMemberRepository, ch.ID, text, entities, &parsed, usedOffsets)

		for _, rule := range rules {
			count := rule.Count
			if count == RuleVariadic {
				count = 50
			}

			parsedCount := 0

			if rule.Type == RuleText {
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

				if rule.TextValidate != nil && !rule.TextValidate(joinedText) {
					return false
				}

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
						if rule.Type == RuleUser {
							if id, err := strconv.ParseInt(tok.text, 10, 64); err == nil {
								if cm, err := r.chatMemberRepository.Get(c.Context, ch.ID, id); err == nil {
									parsed.Users = append(parsed.Users, cm)
									usedOffsets = append(usedOffsets, Offset{tok.start, tok.end})
									parsedCount++
									matched = true
									break
								}
							}
						}

						if rule.Type == RuleDuration {
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

						if rule.Type == RuleDateTime {
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

						if rule.Type == RuleNumber {
							if num, err := strconv.ParseInt(tok.text, 10, 64); err == nil {
								parsed.Numbers = append(parsed.Numbers, num)
								usedOffsets = append(usedOffsets, Offset{tok.start, tok.end})
								parsedCount++
								matched = true
								break
							}
						}
					}

					if !matched && rule.Type == RuleUser && parsedCount == 0 {
						if cm, ok := r.resolveReplyUser(c.Context, ch.ID, c); ok {
							parsed.Users = append(parsed.Users, cm)
							parsedCount++
							matched = true
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

		remaining := getFreeTokens(text, usedOffsets)
		if len(remaining) > 0 {
			return false
		}

		c.Context = context.WithValue(c.Context, commandParsedArgsKey, parsed)

		return true
	}
}

func allRulesAreOptional(rules ...Rule) bool {
	for _, rule := range rules {
		if !rule.Optional {
			return false
		}
	}
	return true
}

func (r *RuleChecker) resolveReplyUser(
	ctx context.Context,
	chatID int64,
	c *botapi.Context,
) (chatmember.ChatMember, bool) {
	m := c.Message()
	if m == nil {
		return chatmember.ChatMember{}, false
	}

	reply := m.ReplyToMessage

	if reply == nil || reply.From == nil {
		return chatmember.ChatMember{}, false
	}

	cm, err := r.chatMemberRepository.Get(ctx, chatID, reply.From.ID)
	if err != nil {
		return chatmember.ChatMember{}, false
	}

	return cm, true
}
