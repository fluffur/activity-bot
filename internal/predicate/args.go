package predicate

import (
	"activity-bot/internal/chatmember"
	"activity-bot/internal/message"
	"activity-bot/internal/middleware/cctx"
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
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
	messageRepository    message.Repository
}

func NewRuleChecker(cmr chatmember.Repository, mr message.Repository) *RuleChecker {
	return &RuleChecker{cmr, mr}
}

func (r *RuleChecker) With(rules ...Rule) botapi.Predicate {
	return func(c *botapi.Context) bool {
		argsMessage := Args(c)

		if argsMessage == nil && !allRulesAreOptionalOrUser(rules...) && len(rules) != 0 {
			return false
		}

		text, entities := argsMessage.TextAndEntities()
		if strings.TrimSpace(text) == "" && !allRulesAreOptionalOrUser(rules...) && len(rules) != 0 {
			spew.Dump("FAIL", "rule user 1")

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
		if len(parsed.Users) == 0 {
			if cm, ok := r.resolveReplyUser(c, ch.ID); ok {
				parsed.Users = append(parsed.Users, cm)
			}
		}

		for _, rule := range rules {
			count := rule.Count
			if count == RuleVariadic {
				count = 50
			}

			parsedCount := 0

			if rule.Type == RuleUser {
				parsedCount = len(parsed.Users)
			}

			if rule.Type == RuleText {
				toks := getFreeTokens(text, usedOffsets)

				if len(toks) == 0 {
					if rule.Optional {
						continue
					}
					spew.Dump("FAIL", "rule user 2")

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
					spew.Dump("FAIL", "rule user 3")

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

					if !matched {
						break
					}
				}
			}

			if parsedCount == 0 && !rule.Optional {
				spew.Dump("FAIL", "rule user 4")

				return false
			}
		}

		remaining := getFreeTokens(text, usedOffsets)
		if len(remaining) > 0 {
			spew.Dump("FAIL", "rule user 5")

			return false
		}

		c.Context = context.WithValue(c.Context, commandParsedArgsKey, parsed)

		return true
	}
}

func allRulesAreOptionalOrUser(rules ...Rule) bool {
	for _, rule := range rules {
		if !rule.Optional && rule.Type != RuleUser {
			return false
		}
	}
	return true
}

func (r *RuleChecker) resolveReplyUser(
	c *botapi.Context,
	chatID int64,
) (chatmember.ChatMember, bool) {
	m := c.Message()
	if m == nil || m.ReplyToMessage == nil {
		return chatmember.ChatMember{}, false
	}

	cm, err := r.messageRepository.GetAuthor(
		c.Context,
		chatID,
		int64(m.ReplyToMessage.MessageID),
	)
	if err == nil {
		return cm, true
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return chatmember.ChatMember{}, false
	}

	reply, err := c.Bot.GetMessage(
		c.Context,
		botapi.ID(chatID),
		m.ReplyToMessage.MessageID,
	)
	if err != nil || reply == nil || reply.From == nil {
		return chatmember.ChatMember{}, false
	}

	cm, err = r.chatMemberRepository.Get(
		c.Context,
		chatID,
		reply.From.ID,
	)
	if err != nil {
		return chatmember.ChatMember{}, false
	}

	return cm, true
}
