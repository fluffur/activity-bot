package predicate

import (
	"activity-bot/internal/cctx"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/gotd/log"

	"github.com/gotd/botapi"
)

type ParsedMessage struct {
	Prefix             string
	Text               string
	Entities           []botapi.MessageEntity
	TrimmedText        string
	LeadingSpacesRunes int
}

func ParseMessage(c *botapi.Context, prefixes []string) (*ParsedMessage, bool) {
	m := c.Message()
	if m == nil {
		return nil, false
	}

	ch, err := cctx.Chat(c)
	if err != nil {
		log.For(c.Bot.Logger()).Error(c, "command ctx chat", log.Error(err))
		return nil, false
	}

	text, entities := m.TextAndEntities()

	if ch.CommandPrefix != "" {
		prefixes = append(prefixes, ch.CommandPrefix)
	}

	prefix := findPrefix(text, prefixes)
	if !ch.AllowPrefixless && prefix == "" {
		return nil, false
	}

	raw := text[len(prefix):]
	rawRunes := []rune(raw)

	leading := 0
	for leading < len(rawRunes) && unicode.IsSpace(rawRunes[leading]) {
		leading++
	}

	trimmed := strings.TrimRightFunc(string(rawRunes[leading:]), unicode.IsSpace)
	if trimmed == "" {
		return nil, false
	}

	return &ParsedMessage{
		Prefix:             prefix,
		Text:               text,
		Entities:           entities,
		TrimmedText:        trimmed,
		LeadingSpacesRunes: leading,
	}, true
}

func BuildArgsMessage(
	m *botapi.Message,
	text string,
	entities []botapi.MessageEntity,
	prefix string,
	leadingSpaces int,
	cmdLenBytes int,
) botapi.Message {
	cmdEndRuneIdx := len([]rune(text[len(prefix)+leadingSpaces : len(prefix)+leadingSpaces+cmdLenBytes]))

	absRuneIdx := len([]rune(prefix)) + leadingSpaces + cmdEndRuneIdx

	runes := []rune(text)
	if absRuneIdx > len(runes) {
		absRuneIdx = len(runes)
	}

	utf16Offset := len(utf16.Encode(runes[:absRuneIdx]))

	var argsEntities []botapi.MessageEntity

	for _, ent := range entities {
		if ent.Offset >= utf16Offset {
			shifted := ent
			shifted.Offset -= utf16Offset
			argsEntities = append(argsEntities, shifted)
		}
	}

	args := botapi.Message{
		MessageID:       m.MessageID,
		MessageThreadID: m.MessageThreadID,
		From:            m.From,
		SenderChat:      m.SenderChat,
		Date:            m.Date,
		Chat:            m.Chat,
		ForwardOrigin:   m.ForwardOrigin,
		ReplyToMessage:  m.ReplyToMessage,
		ViaBot:          m.ViaBot,
		EditDate:        m.EditDate,
	}

	argsText := string(runes[absRuneIdx:])

	if m.Text != "" {
		args.Text = argsText
		args.Entities = argsEntities
	} else {
		args.Caption = argsText
		args.CaptionEntities = argsEntities
	}

	return args
}
