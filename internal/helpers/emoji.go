package helpers

import "fmt"

func CustomEmoji(id int64, originalEmoji string) string {
	return fmt.Sprintf(`<tg-emoji emoji-id="%d">%s</tg-emoji>`, id, originalEmoji)
}

func Line() string {
	middleEmoji := CustomEmoji(5404333313919834615, "↔️")
	middleEmojis := ""
	for range 7 {
		middleEmojis += middleEmoji
	}
	return CustomEmoji(5404805970775792817, "⬅️") + middleEmojis + CustomEmoji(5404631702477757552, "➡️")
}

const (
	DangerEmojiGray  = "5416076321442777828"
	SuccessEmojiGray = "5411197345968701560"
	SuccessEmojiID   = 5411197345968701560
	DangerEmojiID    = 5416076321442777828
)

func SuccessEmoji() string {
	return CustomEmoji(SuccessEmojiID, "✅")
}

func DangerEmoji() string {
	return CustomEmoji(DangerEmojiID, "❌")
}

func StatsEmoji() string {
	return CustomEmoji(5425112292683435471, "📊")
}
