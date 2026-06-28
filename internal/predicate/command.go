package predicate

import (
	"activity-bot/internal/middleware/cctx"
	"context"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/gotd/log"

	"github.com/gotd/botapi"
)

var defaultPrefixes = []string{"!", "/", ".", "фм"}

type argsKey struct{}

var commandArgsKey = argsKey{}

func Args(c *botapi.Context) *botapi.Message {
	if val, ok := c.Context.Value(commandArgsKey).(*botapi.Message); ok {
		return val
	}

	return nil
}
func Command(name string, aliases ...string) botapi.Predicate {
	commands := make([]string, 0, len(aliases)+1)
	commands = append(commands, normalize(name))
	for _, alias := range aliases {
		commands = append(commands, normalize(alias))
	}

	return func(c *botapi.Context) bool {
		m := c.Message()
		if m == nil {
			return false
		}

		ch, err := cctx.Chat(c.Context)
		if err != nil {
			log.For(c.Bot.Logger()).Error(c.Context, "command ctx chat", log.Error(err))
			return false
		}

		text, entities := m.TextAndEntities()

		prefixes := append([]string{}, defaultPrefixes...)
		if ch.CommandPrefix != "" {
			prefixes = append(prefixes, ch.CommandPrefix)
		}

		prefix := findPrefix(text, prefixes)
		if !ch.AllowPrefixless && prefix == "" {
			return false
		}

		rawTextAfterPrefix := text[len(prefix):]
		rawTextAfterPrefixRunes := []rune(rawTextAfterPrefix)

		leadingSpacesRunes := 0
		for leadingSpacesRunes < len(rawTextAfterPrefixRunes) && unicode.IsSpace(rawTextAfterPrefixRunes[leadingSpacesRunes]) {
			leadingSpacesRunes++
		}

		trimmedText := string(rawTextAfterPrefixRunes[leadingSpacesRunes:])
		trimmedText = strings.TrimRightFunc(trimmedText, unicode.IsSpace)

		if trimmedText == "" {
			return false
		}

		self := c.Bot.Self()
		if self == nil {
			return false
		}
		botUsername := strings.ToLower(self.Username)

		for _, cmd := range commands {
			cmdLenBytes, ok := matchCommandAndGetLen(trimmedText, cmd, botUsername)
			if !ok {
				continue
			}

			cmdEndRuneIdxInTrimmed := len([]rune(trimmedText[:cmdLenBytes]))

			prefixRunes := []rune(prefix)
			absRuneIdx := len(prefixRunes) + leadingSpacesRunes + cmdEndRuneIdxInTrimmed

			runes := []rune(text)
			if absRuneIdx > len(runes) {
				absRuneIdx = len(runes)
			}

			utf16Offset := len(utf16.Encode(runes[:absRuneIdx]))

			var argsEntities []botapi.MessageEntity
			for _, ent := range entities {
				if ent.Offset >= utf16Offset {
					shiftedEnt := ent
					shiftedEnt.Offset = ent.Offset - utf16Offset
					argsEntities = append(argsEntities, shiftedEnt)
				}
			}

			argsMessage := &botapi.Message{
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
				argsMessage.Text = argsText
				argsMessage.Entities = argsEntities
			} else {
				argsMessage.Caption = argsText
				argsMessage.CaptionEntities = argsEntities
			}

			c.Context = context.WithValue(c.Context, commandArgsKey, argsMessage)
			return true
		}

		return false
	}
}

func matchCommandAndGetLen(trimmedText string, command string, botUsername string) (int, bool) {
	firstLine := trimmedText
	if idx := strings.Index(trimmedText, "\n"); idx != -1 {
		firstLine = trimmedText[:idx]
	}

	words := strings.Fields(firstLine)
	if len(words) == 0 {
		return 0, false
	}

	cmdWords := strings.Fields(command)
	if len(words) < len(cmdWords) {
		return 0, false
	}

	for i, cmdWord := range cmdWords {
		word := strings.ToLower(words[i])
		last := i == len(cmdWords)-1

		if last {
			if word == cmdWord {
				continue
			}
			if strings.HasPrefix(word, cmdWord+"@") {
				username := strings.TrimPrefix(word, cmdWord+"@")
				if username == botUsername {
					continue
				}
			}
			return 0, false
		}

		if word != cmdWord {
			return 0, false
		}
	}

	pos := len(cmdWords)

	if len(words) > pos && strings.HasPrefix(words[pos], "@") {
		if strings.ToLower(words[pos][1:]) == botUsername {
			pos++
		}
	}

	lastWordToFind := words[pos-1]
	lastWordIdx := strings.Index(strings.ToLower(firstLine), strings.ToLower(lastWordToFind))
	if lastWordIdx == -1 {
		return 0, false
	}

	bytesLen := lastWordIdx + len(lastWordToFind)
	return bytesLen, true
}

func findWordStartRuneIndex(text string, wordIndex int) int {
	runes := []rune(text)
	n := len(runes)

	inWord := false
	wordCount := 0

	for i := 0; i < n; i++ {
		isSpace := unicode.IsSpace(runes[i])
		if !isSpace && !inWord {
			inWord = true
			if wordCount == wordIndex {
				return i
			}
			wordCount++
		} else if isSpace && inWord {
			inWord = false
		}
	}
	return n
}

func findWordEndRuneIndex(text string, wordIndex int) int {
	runes := []rune(text)
	n := len(runes)

	inWord := false
	wordCount := 0

	for i := 0; i < n; i++ {
		isSpace := unicode.IsSpace(runes[i])
		if !isSpace && !inWord {
			inWord = true
			wordCount++
		} else if isSpace && inWord {
			inWord = false
			if wordCount == wordIndex {
				return i
			}
		}
	}
	if inWord && wordCount == wordIndex {
		return n
	}
	return n
}

func parseCommand(
	originalText string,
	command string,
	botUsername string,
) (string, int, bool) {
	words := strings.Fields(originalText)

	if len(words) == 0 {
		return "", 0, false
	}

	cmdWords := strings.Fields(command)

	if len(words) < len(cmdWords) {
		return "", 0, false
	}

	for i, cmdWord := range cmdWords {
		word := strings.ToLower(words[i])

		last := i == len(cmdWords)-1

		if last {
			if word == cmdWord {
				continue
			}

			if strings.HasPrefix(word, cmdWord+"@") {
				username := strings.TrimPrefix(word, cmdWord+"@")

				if username == botUsername {
					continue
				}
			}

			return "", 0, false
		}

		if word != cmdWord {
			return "", 0, false
		}
	}

	pos := len(cmdWords)

	if len(words) > pos && strings.HasPrefix(words[pos], "@") {
		if strings.ToLower(words[pos][1:]) != botUsername {
			return "", 0, false
		}

		pos++
	}

	if len(words) <= pos {
		return "", pos, true
	}

	return strings.Join(words[pos:], " "), pos, true
}

func normalize(s string) string {
	return strings.Join(
		strings.Fields(strings.ToLower(s)),
		" ",
	)
}

func findPrefix(text string, prefixes []string) string {
	textLower := strings.ToLower(text)

	longest := ""

	for _, prefix := range prefixes {
		if strings.HasPrefix(textLower, strings.ToLower(prefix)) &&
			len(prefix) > len(longest) {
			longest = prefix
		}
	}

	return longest
}
