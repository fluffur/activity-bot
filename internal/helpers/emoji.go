package helpers

import (
	"activity-bot/internal/model"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/tg"
	"github.com/makeworld-the-better-one/go-isemoji"
	"github.com/rivo/uniseg"
)

type span struct {
	start int
	end   int
}

func ExtractEmoji(text string, entities []tg.MessageEntityClass) model.Emojis {
	var result model.Emojis

	runes := []rune(text)
	utf16Codes := utf16.Encode(runes)

	entityMap := make(map[int]*tg.MessageEntityCustomEmoji)

	for _, e := range entities {
		if v, ok := e.(*tg.MessageEntityCustomEmoji); ok {
			entityMap[v.Offset] = v
		}
	}

	g := uniseg.NewGraphemes(text)

	pos := 0

	for g.Next() {
		part := g.Str()

		partUTF16 := utf16.Encode([]rune(part))
		size := len(partUTF16)

		if ent, ok := entityMap[pos]; ok {

			start := ent.Offset
			end := ent.Offset + ent.Length

			if end <= len(utf16Codes) {
				sub := utf16Codes[start:end]

				result = append(result, model.Emoji{
					Type: model.EmojiTypeCustom,
					ID:   ent.DocumentID,
					Char: string(utf16.Decode(sub)),
				})

				pos = end
				continue
			}
		}

		if isemoji.IsEmoji(part) {
			result = append(result, model.Emoji{
				Type: model.EmojiTypeUnicode,
				Char: part,
			})
		}

		pos += size
	}

	return result
}

func isUsed(start, end int, used []span) bool {
	for _, s := range used {
		if start < s.end && end > s.start {
			return true
		}
	}
	return false
}

func DisplayEmoji(eb *entity.Builder, emojis model.Emojis) {
	for _, emoji := range emojis {
		switch emoji.Type {
		case model.EmojiTypeCustom:
			eb.CustomEmoji(emoji.Char, emoji.ID)
		case model.EmojiTypeUnicode:
			eb.Plain(emoji.Char)

		}
	}
}

func MentionEmoji(eb *entity.Builder, user model.User, emojis model.Emojis) {
	var hasNormalEmoji bool
	for _, emoji := range emojis {
		switch emoji.Type {
		case model.EmojiTypeCustom:
			eb.CustomEmoji(emoji.Char, emoji.ID)
		case model.EmojiTypeUnicode:
			if !hasNormalEmoji {
				WriteMention(eb, user.ID, emoji.Char)
				hasNormalEmoji = true
			} else {
				eb.Plain(emoji.Char)
			}
		}
	}
	if !hasNormalEmoji {
		WriteMention(eb, user.ID, "​")
	}
}

func CustomEmoji(id string, originalEmoji string) string {
	return fmt.Sprintf(`<tg-emoji emoji-id="%s">%s</tg-emoji>`, id, originalEmoji)
}

var tgEmojiRegex = regexp.MustCompile(`<tg-emoji[^>]*>.*?</tg-emoji>`)

func ParseEmojis(input string) []string {
	var result []string

	for len(input) > 0 {

		if strings.HasPrefix(input, "<tg-emoji") {
			loc := tgEmojiRegex.FindStringIndex(input)
			if loc != nil {
				result = append(result, input[loc[0]:loc[1]])
				input = input[loc[1]:]
				continue
			}
		}

		g := uniseg.NewGraphemes(input)
		if g.Next() {
			part := g.Str()

			if isemoji.IsEmoji(part) {
				result = append(result, part)
			}

			input = input[len(part):]
			continue
		}

		break
	}

	return result
}

func NewbieEmoji() string {
	return CustomEmoji("5235782484939012025", "🐣")
}

func TotalEmoji() string {
	return CustomEmoji("5870753782874246579", "📝")
}

func RestEmoji() string {
	return CustomEmoji("5235961361736956044", "💤")
}

func Line() string {
	return ""
	//middleEmoji := CustomEmoji("5404333313919834615", "↔️")
	//middleEmojis := ""
	//for range 7 {
	//	middleEmojis += middleEmoji
	//}
	//return CustomEmoji("5404805970775792817", "⬅️") + middleEmojis + CustomEmoji("5404631702477757552", "➡️")
}

const (
	DangerEmojiGray  = "5416076321442777828"
	SuccessEmojiGray = "5411197345968701560"
	SuccessEmojiID   = "5411197345968701560"
	DangerEmojiID    = "5416076321442777828"
)

func SuccessEmoji() string {
	return CustomEmoji(SuccessEmojiID, "✅")
}

func DangerEmoji() string {
	return CustomEmoji(DangerEmojiID, "❌")
}

func StatsEmoji() string {
	return CustomEmoji("5258391025281408576", "📊")
}

func WriteCustomEmoji(eb *entity.Builder, id string, originalEmoji string) {
	docID, _ := strconv.ParseInt(id, 10, 64)
	eb.CustomEmoji(originalEmoji, docID)
}

func WriteNewbieEmoji(eb *entity.Builder) {
	WriteCustomEmoji(eb, "5235782484939012025", "🐣")
}

func WriteTotalEmoji(eb *entity.Builder) {
	WriteCustomEmoji(eb, "5870753782874246579", "📝")
}

func WriteRestEmoji(eb *entity.Builder) {
	WriteCustomEmoji(eb, "5235961361736956044", "💤")
}

func WriteSuccessEmoji(eb *entity.Builder) {
	WriteCustomEmoji(eb, SuccessEmojiID, "✅")
}

func WriteDangerEmoji(eb *entity.Builder) {
	WriteCustomEmoji(eb, DangerEmojiID, "❌")
}

func WriteStatsEmoji(eb *entity.Builder) {
	WriteCustomEmoji(eb, "5258391025281408576", "📊")
}

func WriteEmoji(eb *entity.Builder, emoji string, emojis model.Emojis) {
	if len(emojis) == 0 {
		eb.Plain(emoji)
	}

	for _, e := range emojis {
		eb.CustomEmoji(e.Char, e.ID)
	}
}
