package predicate

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/message"
	"activity-bot/internal/rule"
	"database/sql"
	"errors"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/gotd/log"

	"github.com/gotd/botapi"
)

func NoArgs() botapi.Predicate {
	return func(c *botapi.Context) bool {
		args, err := cctx.ArgsMessage(c)
		return err != nil || strings.TrimSpace(args.Text) == ""
	}
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

func (r *RuleChecker) With(rules ...rule.Rule) botapi.Predicate {
	return func(c *botapi.Context) bool {
		argsMessage, err := cctx.ArgsMessage(c)

		if err != nil && !allRulesAreOptionalOrUser(rules...) && len(rules) != 0 {
			return false
		}

		text, entities := argsMessage.TextAndEntities()
		if strings.TrimSpace(text) == "" && !allRulesAreOptionalOrUser(rules...) && len(rules) != 0 {
			spew.Dump("FAIL", "rule user 1")

			return false
		}

		ch, err := cctx.Chat(c)
		if err != nil {
			log.For(c.Bot.Logger()).Error(c, "with args: chat ctx missing", log.Error(err))

			return false
		}

		var usedOffsets []Offset

		parsed := cctx.ParsedArgs{}

		usedOffsets = resolveUserEntities(c, r.chatMemberRepository, ch.ID, text, entities, &parsed, usedOffsets)
		if len(parsed.Users) == 0 {
			if cm, ok := r.resolveReplyUser(c, ch.ID); ok {
				parsed.Users = append(parsed.Users, cm)
			}
		}

		for _, rul := range rules {
			count := rul.CountArgs
			if count == rule.RuleVariadic {
				count = 50
			}

			parsedCount := 0

			if rul.Type == rule.RuleUser {
				parsedCount = len(parsed.Users)
			}

			if rul.Type == rule.RuleText {
				toks := getFreeTokens(text, usedOffsets)

				if len(toks) == 0 {
					if rul.IsOptional {
						continue
					}

					spew.Dump("FAIL", "rule user 2")

					return false
				}

				var textParts []string

				if rul.OnNextRow {
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

				if rul.TextValidate != nil && !rul.TextValidate(joinedText) {
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
						if rul.Type == rule.RuleUser {
							if id, err := strconv.ParseInt(tok.text, 10, 64); err == nil {
								if cm, err := r.chatMemberRepository.Get(c, ch.ID, id); err == nil {
									parsed.Users = append(parsed.Users, cm)
									usedOffsets = append(usedOffsets, Offset{tok.start, tok.end})
									parsedCount++

									matched = true

									break
								}
							}
						}

						if rul.Type == rule.RuleDuration {
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

						if rul.Type == rule.RuleDateTime {
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

						if rul.Type == rule.RuleDateTimeOrDuration {
							if d, consumed, ok := tryParseAdvancedDuration(toks[idx:]); ok {
								parsed.Durations = append(parsed.Durations, d)

								for k := 0; k < consumed; k++ {
									usedOffsets = append(
										usedOffsets,
										Offset{toks[idx+k].start, toks[idx+k].end},
									)
								}

								parsedCount++

								matched = true

								break
							}

							if t, consumed, ok := tryParseAdvancedDateTime(toks[idx:]); ok {
								parsed.DateTimes = append(parsed.DateTimes, t)

								for k := 0; k < consumed; k++ {
									usedOffsets = append(
										usedOffsets,
										Offset{toks[idx+k].start, toks[idx+k].end},
									)
								}

								parsedCount++

								matched = true

								break
							}
						}

						if rul.Type == rule.RuleNumber {
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

			if parsedCount == 0 && !rul.IsOptional {
				spew.Dump("FAIL", "rule user 4", parsedCount, rul.IsOptional)

				return false
			}
		}

		remaining := getFreeTokens(text, usedOffsets)
		if len(remaining) > 0 {
			spew.Dump("FAIL", "rule user 5", text, usedOffsets)

			return false
		}

		c.Context = cctx.WithParsedArgs(c.Context, parsed)

		return true
	}
}

func allRulesAreOptionalOrUser(rules ...rule.Rule) bool {
	for _, r := range rules {
		if !r.IsOptional && r.Type != rule.RuleUser {
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
		c,
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
		c,
		botapi.ID(chatID),
		m.ReplyToMessage.MessageID,
	)
	if err != nil || reply == nil || reply.From == nil {
		return chatmember.ChatMember{}, false
	}

	cm, err = r.chatMemberRepository.Get(
		c,
		chatID,
		reply.From.ID,
	)
	if err != nil {
		return chatmember.ChatMember{}, false
	}

	return cm, true
}
