package helpers

import (
	"activity-bot/internal/model"
	"fmt"
	"strconv"
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

const (
	SuccessEmojiID = "5411197345968701560"
	DangerEmojiID  = "5416076321442777828"
)

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
