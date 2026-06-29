package summon

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"strings"
	"time"

	"github.com/gotd/log"
	"golang.org/x/time/rate"

	"github.com/gotd/botapi"
)

func TextFromArgs(ch chat.Chat, args *botapi.Message) string {
	var text string
	if args != nil {
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

func (h *Handler) Summon(c *botapi.Context, text string, msgID int, ch chat.Chat, cms []chatmember.ChatMember) error {
	if _, loaded := h.activeSummons.LoadOrStore(ch.ID, struct{}{}); loaded {
		_, err := c.Reply(h.translator.T(ch.Lang, i18n.Cmd.Summon.AlreadyRunning))
		return err
	}

	mentions := BuildMentions(cms, ch.MentionTypes)

	sep := mentionSeparator(ch.MentionTypes)

	perMsg := int(ch.MentionsPerMessage)
	if perMsg <= 0 {
		perMsg = len(mentions)
	}

	msgs := BuildMentionMessages(text, mentions, perMsg, sep)

	go func() {
		defer h.activeSummons.Delete(ch.ID)

		if err := SendMessages(
			c,
			h.translator,
			ch.Lang,
			ch.ID,
			msgID,
			msgs,
		); err != nil {
			log.For(c.Bot.Logger()).Error(c.Context, "send messages", log.Error(err))
		}
	}()

	return nil
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
	t *i18n.Translator,
	lang string,
	chatID int64,
	msgID int,
	messages []string,
) error {
	msg, err := c.Bot.GetMessage(c.Background(), botapi.ID(chatID), msgID)
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

		var err error

		if photoID != "" {
			_, err = c.Bot.SendPhoto(
				c.Background(),
				botapi.ID(chatID),
				botapi.FileID(photoID),
				text,
				botapi.WithParseMode(botapi.ParseModeHTML),
				botapi.ReplyTo(msgID),
			)
		} else {
			_, err = c.Bot.SendMessage(
				c.Background(),
				botapi.ID(chatID),
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
		botapi.ID(chatID),
		t.T(lang, i18n.Cmd.Summon.Completed),
		botapi.WithParseMode(botapi.ParseModeHTML),
		botapi.DisableWebPagePreview(),
	)

	return err
}

func RenderMention(cm chatmember.ChatMember, mentionTypes chat.MentionTypes) string {
	hasEmoji := mentionTypes.Has(chat.MentionEmoji)
	hasRole := mentionTypes.Has(chat.MentionRole)
	hasName := mentionTypes.Has(chat.MentionName)

	result := ""

	if hasEmoji && cm.AnyEmoji() != "" {
		result += cm.AnyEmoji() + " "
	}

	if hasRole && hasName && cm.Tag != "" {
		result += fmt.Sprintf("%s (%s)", cm.Tag, cm.User.FullName())
	} else if hasRole && cm.Tag != "" {
		result += cm.Tag
	} else if hasName {
		result += cm.User.FullName()
	}

	if hasEmoji && !hasRole && !hasName {
		result += "​"
	}

	if strings.TrimSpace(result) == "" {
		result = "​"
	}

	return tghtml.UserMention(cm.User.ID, result)
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

func mentionSeparator(mt chat.MentionTypes) string {
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

	if !ok || session.State != StateAwaitConfirmation {
		return nil, chat.Chat{}, nil
	}

	ch, err := h.chatService.Get(c.Context, session.Data.ChatID)
	if err != nil {
		return nil, chat.Chat{}, fmt.Errorf("get chat: %w", err)
	}

	sender := c.Sender()
	if sender == nil {
		return nil, ch, nil
	}

	if session.Data.UserID != sender.ID {
		return nil, ch, c.AnswerCallback(
			botapi.WithCallbackText(
				h.translator.T(
					ch.Lang,
					i18n.Cmd.Summon.Confirm.OnlyInitiatorConfirm,
				),
			),
		)
	}

	return &session.Data, ch, nil
}
