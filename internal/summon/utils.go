package summon

import (
	"activity-bot/internal/cctx"
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/log"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
	"golang.org/x/time/rate"

	"github.com/gotd/botapi"
)

func TextFromArgs(ch chat.Chat, args botapi.Message) string {
	var text string

	if args.OriginalTextHTML() != "" {
		text = ReplaceMentions(args.OriginalTextHTML(), args.Entities)
	}

	if strings.TrimSpace(text) == "" {
		text = ch.WelcomeCallMessage
	}

	if strings.TrimSpace(text) == "" && ch.MentionTypes == 0 {
		text = "ㅤ"
	}

	return text
}

func (h *Handler) Summon(
	c *botapi.Context,
	text string,
	msgID int,
	ch chat.Chat,
	cms []chatmember.ChatMember,
) error {
	chatID, ok := c.Chat()
	if !ok {
		return nil
	}

	if _, loaded := h.activeSummons.LoadOrStore(ch.ID, struct{}{}); loaded {
		loc := cctx.MustLocalizer(c)

		_, err := c.Reply(
			loc.T(i18n.Cmd.Summon.AlreadyRunning, nil),
		)

		return err
	}

	perMsg := int(ch.MentionsPerMessage)
	if perMsg <= 0 {
		perMsg = len(cms)
	}

	groups := chunk(cms, perMsg)

	go func() {
		defer h.activeSummons.Delete(ch.ID)

		if err := SendMessagesNew(
			c,
			cctx.MustLocalizer(c),
			chatID,
			msgID,
			text,
			ch.MentionTypes,
			groups,
		); err != nil {
			log.For(c.Bot.Logger()).Error(c, "send messages", log.Error(err))
		}
	}()

	return nil
}

var customEmojiRegexp = regexp.MustCompile(
	`<tg-emoji emoji-id="(\d+)">(.*?)</tg-emoji>`,
)

func ParseCustomEmoji(s string) (int64, string) {
	m := customEmojiRegexp.FindStringSubmatch(s)
	if len(m) != 3 {
		return 0, ""
	}

	id, _ := strconv.ParseInt(m[1], 10, 64)

	return id, m[2]
}

func BuildMentionMessage(
	eb *entity.Builder,
	text string,
	members []chatmember.ChatMember,
	mentionTypes chat.MentionTypes,
) {
	if strings.TrimSpace(text) != "" && text != "ㅤ" {
		eb.Plain(text)

		if mentionTypes != 0 {
			eb.Plain("\n\n")
		}
	}

	sep := " "
	if mentionTypes.Has(chat.MentionName) || mentionTypes.Has(chat.MentionRole) {
		sep = ", "
	}

	for i, cm := range members {
		RenderMentionEB(eb, cm, mentionTypes)

		if i != len(members)-1 {
			eb.Plain(sep)
		}
	}

	eb.Plain("ㅤ")
}

func BuildMentionMessages(
	text string,
	mentions []string,
	perMsg int,
	sep string,
) []string {
	if perMsg <= 0 {
		perMsg = len(mentions)
	}

	groups := chunk(mentions, perMsg)

	result := make([]string, 0, len(groups))

	for _, g := range groups {
		msg := text

		if strings.TrimSpace(msg) != "" && msg != "ㅤ" {
			msg += "\n\n"
		}

		msg += strings.Join(g, sep)
		result = append(result, msg)
	}

	return result
}

func SendMessages(
	c *botapi.Context,
	loc *i18n.Localizer,
	chatID botapi.ChatID,
	msgID int,
	messages []string,
) error {
	msg, err := c.Bot.GetMessage(c.Background(), chatID, msgID)
	if err != nil {
		return err
	}

	var photoID string

	if msg.Photo != nil && len(msg.Photo) > 0 {
		photoID = msg.Photo[len(msg.Photo)-1].FileID
	}

	chatLimiter := rate.NewLimiter(rate.Every(1500*time.Microsecond), 1)

	for _, text := range messages {
		if err := chatLimiter.Wait(c.Background()); err != nil {
			return fmt.Errorf("send summon messages: %w", err)
		}

		if photoID != "" {
			_, err = c.Bot.SendPhoto(
				c.Background(),
				chatID,
				botapi.FileID(photoID),
				text,
				botapi.WithParseMode(botapi.ParseModeHTML),
				botapi.ReplyTo(msgID),
			)
		} else {
			_, err = c.Bot.SendMessage(
				c.Background(),
				chatID,
				text,
				botapi.WithParseMode(botapi.ParseModeHTML),
				botapi.DisableWebPagePreview(),
				botapi.ReplyTo(msgID),
			)
		}

		if err != nil {
			return err
		}
	}

	if err := chatLimiter.Wait(c.Background()); err != nil {
		return fmt.Errorf("send summon last msg: %w", err)
	}

	_, err = c.Bot.SendMessage(
		c.Background(),
		chatID,
		loc.T(i18n.Cmd.Summon.Completed, nil),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}

func SendMessagesNew(
	c *botapi.Context,
	loc *i18n.Localizer,
	chatID botapi.ChatID,
	msgID int,
	text string,
	mentionTypes chat.MentionTypes,
	groups [][]chatmember.ChatMember,
) error {
	_, err := c.Bot.GetMessage(c.Background(), chatID, msgID)
	if err != nil {
		return err
	}

	//var photoID string
	//
	//if msg.Photo != nil && len(msg.Photo) > 0 {
	//	photoID = msg.Photo[len(msg.Photo)-1].FileID
	//}

	limiter := rate.NewLimiter(rate.Every(1500*time.Microsecond), 1)
	peer, err := c.Bot.Peers().ResolveTDLibID(
		c.Background(),
		constant.TDLibPeerID(chatID.(botapi.ChatIDInt)),
	)
	if err != nil {
		return err
	}
	inputPeer := peer.InputPeer()

	if err != nil {
		return err
	}
	for _, group := range groups {
		if err := limiter.Wait(c.Background()); err != nil {
			return err
		}

		eb := &entity.Builder{}

		BuildMentionMessage(
			eb,
			text,
			group,
			mentionTypes,
		)

		finalText, entities := eb.Complete()

		_, err = c.Bot.Raw().MessagesSendMessage(
			c.Background(),
			&tg.MessagesSendMessageRequest{
				Peer:     inputPeer,
				Message:  finalText,
				Entities: entities,
				RandomID: rand.Int63(),
				ReplyTo: &tg.InputReplyToMessage{
					ReplyToMsgID: msgID,
				},
			},
		)

		if err != nil {
			return err
		}
	}

	if err := limiter.Wait(c.Background()); err != nil {
		return err
	}

	_, err = c.Bot.SendMessage(
		c.Background(),
		chatID,
		loc.T(i18n.Cmd.Summon.Completed, nil),
	)

	return err
}

func RenderMention(cm chatmember.ChatMember, mentionTypes chat.MentionTypes) string {
	hasEmoji := mentionTypes.Has(chat.MentionEmoji)
	hasRole := mentionTypes.Has(chat.MentionRole)
	hasName := mentionTypes.Has(chat.MentionName)

	var text strings.Builder

	role := cm.Role()
	if role == "" {
		role = cm.User.FullName()
	}

	switch {
	case hasRole && hasName:
		if role == cm.User.FullName() {
			text.WriteString(role)
		} else {
			text.WriteString(fmt.Sprintf("%s (%s)", role, cm.User.FullName()))
		}

	case hasRole:
		text.WriteString(role)

	case hasName:
		text.WriteString(cm.User.FullName())
	}

	mentionText := strings.TrimSpace(text.String())

	if mentionText == "" {
		if hasEmoji {
			if emoji := cm.AnyEmoji(); emoji != "" && !strings.Contains(emoji, "<tg-emoji") {
				return tghtml.UserMention(cm.ID(), emoji)
			}
		}

		mentionText = "\u200B"
	}

	mention := tghtml.UserMention(cm.ID(), mentionText)

	if hasEmoji {
		if emoji := cm.AnyEmoji(); emoji != "" {
			return emoji + " " + mention
		}
	}

	return mention
}

func RenderMentionEB(
	eb *entity.Builder,
	cm chatmember.ChatMember,
	mentionTypes chat.MentionTypes,
) {
	hasEmoji := mentionTypes.Has(chat.MentionEmoji)
	hasRole := mentionTypes.Has(chat.MentionRole)
	hasName := mentionTypes.Has(chat.MentionName)

	role := cm.Role()
	if role == "" {
		role = cm.User.FullName()
	}

	var title string

	switch {
	case hasRole && hasName:
		if role == cm.User.FullName() {
			title = role
		} else {
			title = fmt.Sprintf("%s (%s)", role, cm.User.FullName())
		}

	case hasRole:
		title = role

	case hasName:
		title = cm.User.FullName()
	}

	if !hasEmoji {
		writeMention(eb, cm.ID(), title)
		return
	}

	emoji := cm.AnyEmoji()

	if emoji == "" {
		writeMention(eb, cm.ID(), title)
		return
	}

	if strings.Contains(emoji, "<tg-emoji") {
		id, char := ParseCustomEmoji(emoji)

		eb.CustomEmoji(char, id)

		if title != "" {
			writeMention(eb, cm.ID(), title)
		} else {
			writeMention(eb, cm.ID(), "\u200B")
		}

		return
	}

	// unicode emoji

	if title == "" {
		writeMention(eb, cm.ID(), emoji)
		return
	}

	eb.Plain(emoji)
	writeMention(eb, cm.ID(), title)
}

func writeMention(eb *entity.Builder, id int64, text string) {
	if text == "" {
		text = "\u200B"
	}

	eb.MentionName(text, &tg.InputUser{
		UserID: id,
	})
}

func ReplaceMentions(text string, entities []botapi.MessageEntity) string {
	result := text

	for i := len(entities) - 1; i >= 0; i-- {
		e := entities[i]

		start := utf16ToRuneIndex(text, e.Offset)
		end := utf16ToRuneIndex(text, e.Offset+e.Length)

		switch e.Type {
		case "mention":
			username := text[start:end]
			replacement := fmt.Sprintf(
				`<a href="https://t.me/%s">%s</a>`,
				username[1:],
				username,
			)

			result = result[:start] + replacement + result[end:]

		case "text_mention":
			if e.User == nil {
				continue
			}

			name := text[start:end]
			replacement := fmt.Sprintf(
				`<a href="tg://user?id=%d">%s</a>`,
				e.User.ID,
				name,
			)

			result = result[:start] + replacement + result[end:]
		default:
		}
	}

	return result
}

func BuildMentions(
	cms []chatmember.ChatMember,
	mentionTypes chat.MentionTypes,
) []string {
	mentions := make([]string, len(cms))

	for i, cm := range cms {
		mentions[i] = RenderMention(cm, mentionTypes)
	}

	return mentions
}

func check(enabled bool) string {
	if enabled {
		return "✅"
	}

	return ""
}

func MentionSeparator(mt chat.MentionTypes) string {
	if mt.Has(chat.MentionEmoji) && !mt.Has(chat.MentionRole) && !mt.Has(chat.MentionName) || mt == 0 {
		return " "
	}

	return ", "
}

func utf16ToRuneIndex(s string, utf16Pos int) int {
	count := 0

	for i, r := range s {
		if count >= utf16Pos {
			return i
		}

		if r > 0xFFFF {
			count += 2
		} else {
			count++
		}
	}

	return len(s)
}

func chunk[T any](items []T, size int) [][]T {
	if size <= 0 {
		return [][]T{items}
	}

	var chunks [][]T

	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}

		chunks = append(chunks, items[i:end])
	}

	return chunks
}

func (h *Handler) summonSession(
	c *botapi.Context,
) (*StateData, chat.Chat, error) {
	session, ok, err := h.summonFSM.Get(c)
	if err != nil {
		return nil, chat.Chat{}, err
	}

	if !ok {
		return nil, chat.Chat{}, nil
	}

	ch, err := h.chatService.Get(c, session.Data.ChatID)
	if err != nil {
		return nil, chat.Chat{}, fmt.Errorf("get chat: %w", err)
	}

	sender := c.Sender()
	if sender == nil {
		return nil, ch, nil
	}

	if session.Data.UserID != sender.ID {
		loc := cctx.MustLocalizer(c)

		return nil, ch, c.AnswerCallback(
			botapi.WithCallbackText(
				loc.T(i18n.Cmd.Summon.Confirm.OnlyInitiatorConfirm, nil),
			),
		)
	}

	return &session.Data, ch, nil
}
