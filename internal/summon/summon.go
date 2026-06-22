package summon

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/middleware/cctx"
	"activity-bot/internal/predicate"
	"activity-bot/internal/utils/tghtml"
	"fmt"
	"log"
	"strings"

	"github.com/gotd/botapi"
)

func (h *Handler) Summon(c *botapi.Context) error {
	ch, err := cctx.Chat(c.Context)
	if err != nil {
		return fmt.Errorf("summon chat: %w", err)
	}

	args := predicate.Args(c)

	var text string
	if args != nil {
		text = ReplaceMentions(args.OriginalTextHTML(), args.Entities)
	}

	if strings.TrimSpace(text) == "" {
		text = ch.WelcomeCallMessage
	}
	if strings.TrimSpace(text) == "" {
		text = "ㅤ"
	}

	cms, err := h.chatMemberService.ListHumanChatMembers(c.Context, ch.ID)
	if err != nil {
		return fmt.Errorf("summon cms list: %w", err)
	}

	mentions := BuildMentions(cms, ch.MentionTypes)

	sep := mentionSeparator(ch.MentionTypes)

	perMsg := int(ch.MentionsPerMessage)
	if perMsg <= 0 {
		perMsg = len(mentions)
	}

	msgs := BuildMentionMessages(text, mentions, perMsg, sep)

	return SendMessages(c, ch.ID, msgs)
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

		if strings.TrimSpace(msg) != "" {
			msg += "\n\n"
		}

		msg += strings.Join(g, sep)
		result = append(result, msg)
	}

	return result
}

func SendMessages(
	c *botapi.Context,
	chatID int64,
	messages []string,
) error {

	msg := c.Update.Message

	var photoID string
	if msg.Photo != nil && len(msg.Photo) > 0 {
		photoID = msg.Photo[len(msg.Photo)-1].FileID
	}

	for _, text := range messages {

		var err error

		if photoID != "" {
			_, err = c.Bot.SendPhoto(
				c.Context,
				botapi.ID(chatID),
				botapi.FileID(photoID),
				text,
				botapi.WithParseMode(botapi.ParseModeHTML),
			)
		} else {
			_, err = c.Reply(
				text,
				botapi.WithParseMode(botapi.ParseModeHTML),
				botapi.DisableWebPagePreview(),
			)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func RenderMention(cm chatmember.ChatMember, mentionTypes chat.MentionTypes) string {
	hasEmoji := mentionTypes.Has(chat.MentionEmoji)
	hasRole := mentionTypes.Has(chat.MentionRole)
	hasName := mentionTypes.Has(chat.MentionName)

	result := ""

	if hasEmoji && cm.AnyEmoji() != "" {
		result += cm.AnyEmoji()
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

	log.Printf("result %q", result)
	log.Println(hasEmoji, hasRole, hasName)
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
