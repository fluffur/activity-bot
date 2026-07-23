package predicate

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/message"
	"activity-bot/internal/rule"
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"unicode"

	"github.com/gotd/log"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"

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

		replyUser, hasReply := r.resolveReplyUser(c, ch.ID)
		replyUsed := false
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

					return false
				}

				var joinedText string

				if strings.Contains(text, "\n") {
					firstTokStart := toks[0].start
					startIdx := 0
					for _, offset := range usedOffsets {
						if offset.End > startIdx && offset.End <= firstTokStart {
							startIdx = offset.End
						}
					}

					joinedText = text[startIdx:]
					usedOffsets = append(usedOffsets, Offset{startIdx, len(text)})
					joinedText = strings.Trim(joinedText, " \t\r")
				} else {
					var parts []string
					for _, tok := range toks {
						parts = append(parts, tok.text)
						usedOffsets = append(usedOffsets, Offset{tok.start, tok.end})
					}
					joinedText = strings.Join(parts, " ")
				}

				if rul.TextValidate != nil && !rul.TextValidate(joinedText) {
					return false
				}

				if joinedText != "" {
					parsed.Texts = append(parsed.Texts, joinedText)
					parsedCount++
				}
			} else {
				for i := 0; i < count; i++ {
					matched := false

					if rul.Type == rule.RuleUser && hasReply && !replyUsed {
						parsed.Users = append(parsed.Users, replyUser)
						replyUsed = true
						parsedCount++
						continue
					}

					toks := getFreeTokens(text, usedOffsets)
					if len(toks) == 0 {
						break
					}

					for idx, tok := range toks {
						if rul.Type == rule.RuleUser {
							if cm, consumed, ok := r.resolveUserToken(c, ch.ID, toks[idx:]); ok {
								parsed.Users = append(parsed.Users, cm)

								for k := 0; k < consumed; k++ {
									usedOffsets = append(usedOffsets, Offset{
										Start: toks[idx+k].start,
										End:   toks[idx+k].end,
									})
								}

								parsedCount++
								matched = true
								break
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
									usedOffsets = append(usedOffsets, Offset{
										Start: toks[idx+k].start,
										End:   toks[idx+k].end,
									})
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
				return false
			}
		}

		remaining := getFreeTokens(text, usedOffsets)
		if len(remaining) > 0 {
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

	peer, err := c.Bot.Peers().ResolveTDLibID(
		c,
		constant.TDLibPeerID(chatID),
	)
	if err != nil {
		return chatmember.ChatMember{}, false
	}

	channel, ok := peer.InputPeer().(*tg.InputPeerChannel)
	if !ok {
		return chatmember.ChatMember{}, false
	}

	msgs, err := c.Bot.Raw().ChannelsGetMessages(
		c,
		&tg.ChannelsGetMessagesRequest{
			Channel: &tg.InputChannel{
				ChannelID:  channel.ChannelID,
				AccessHash: channel.AccessHash,
			},
			ID: []tg.InputMessageClass{
				&tg.InputMessageID{
					ID: m.ReplyToMessage.MessageID,
				},
			},
		},
	)
	if err != nil {
		return chatmember.ChatMember{}, false
	}

	messages, ok := msgs.(*tg.MessagesChannelMessages)
	if !ok || len(messages.Messages) == 0 {
		return chatmember.ChatMember{}, false
	}

	var userID int64

	switch msg := messages.Messages[0].(type) {
	case *tg.Message:
		if from, ok := msg.FromID.(*tg.PeerUser); ok {
			userID = from.UserID
		}

	case *tg.MessageService:
		switch action := msg.Action.(type) {
		case *tg.MessageActionChatAddUser:
			if len(action.Users) > 0 {
				userID = action.Users[0]
			}

		case *tg.MessageActionChatJoinedByLink:
			if from, ok := msg.FromID.(*tg.PeerUser); ok {
				userID = from.UserID
			}

		case *tg.MessageActionChatJoinedByRequest:
			if from, ok := msg.FromID.(*tg.PeerUser); ok {
				userID = from.UserID
			}

		case *tg.MessageActionChatDeleteUser:
			userID = action.UserID
		}
	}

	if userID == 0 {
		return chatmember.ChatMember{}, false
	}

	cm, err = r.chatMemberRepository.Get(c, chatID, userID)
	if err != nil {
		return chatmember.ChatMember{}, false
	}

	return cm, true
}

func (r *RuleChecker) resolveUserToken(
	ctx context.Context,
	chatID int64,
	tokens []token,
) (chatmember.ChatMember, int, bool) {
	if len(tokens) == 0 {
		return chatmember.ChatMember{}, 0, false
	}

	if id, err := strconv.ParseInt(tokens[0].text, 10, 64); err == nil {
		cm, err := r.chatMemberRepository.Get(ctx, chatID, id)
		if err == nil {
			return cm, 1, true
		}
	}

	if cm, consumed, ok := r.resolveUserTag(ctx, chatID, tokens); ok {
		return cm, consumed, true
	}

	return chatmember.ChatMember{}, 0, false
}

func (r *RuleChecker) resolveUserTag(
	ctx context.Context,
	chatID int64,
	tokens []token,
) (chatmember.ChatMember, int, bool) {
	members, err := r.chatMemberRepository.List(ctx, chatmember.Filter{
		ChatID: chatID,
	})
	if err != nil {
		return chatmember.ChatMember{}, 0, false
	}

	const maxWords = 3

	maxTokens := len(tokens)
	if maxTokens > maxWords {
		maxTokens = maxWords
	}

	for words := maxTokens; words >= 1; words-- {
		var (
			b        strings.Builder
			consumed int
		)

		for i := 0; i < words; i++ {
			part := NormalizeTag(tokens[i].text)
			if part == "" {
				continue
			}

			if consumed > 0 {
				b.WriteByte(' ')
			}

			b.WriteString(part)
			consumed++
		}

		if consumed == 0 {
			continue
		}

		query := b.String()

		var leftMember *chatmember.ChatMember

		for _, member := range members {
			if member.Tag == "" {
				continue
			}

			if !tagContains(NormalizeTag(member.Tag), query) {
				continue
			}

			if member.LeftAt.IsZero() {
				return member, consumed, true
			}

			if leftMember == nil {
				leftMember = new(chatmember.ChatMember)
				*leftMember = member
			}
		}

		if leftMember != nil {
			return *leftMember, consumed, true
		}
	}

	return chatmember.ChatMember{}, 0, false
}

func tagContains(tag, query string) bool {
	tagWords := strings.Fields(tag)
	queryWords := strings.Fields(query)

	if len(queryWords) == 0 || len(queryWords) > len(tagWords) {
		return false
	}

	for i := 0; i <= len(tagWords)-len(queryWords); i++ {
		ok := true

		for j := range queryWords {
			if tagWords[i+j] != queryWords[j] {
				ok = false
				break
			}
		}

		if ok {
			return true
		}
	}

	return false
}

func NormalizeTag(s string) string {
	s = strings.ToLower(s)

	var b strings.Builder
	lastSpace := true
	started := false

	for _, r := range s {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(r)
			lastSpace = false
			started = true

		case started && (r == '.' || r == '-'):
			b.WriteRune(r)
			lastSpace = false

		case started && unicode.IsSpace(r):
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
			}

		case started:
			return strings.TrimSpace(b.String())

		default:
			continue
		}
	}

	return strings.TrimSpace(b.String())
}
