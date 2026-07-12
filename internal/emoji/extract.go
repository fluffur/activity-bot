package emoji

import (
	"regexp"

	"github.com/makeworld-the-better-one/go-isemoji"
	"github.com/rivo/uniseg"
)

var tgEmojiRE = regexp.MustCompile(
	`<tg-emoji\s+emoji-id="\d+">.*?</tg-emoji>`,
)

func Extract(s string) (emojis []string) {
	for {
		loc := tgEmojiRE.FindStringIndex(s)
		if loc == nil {
			emojis = append(emojis, extractUnicode(s)...)
			break
		}

		emojis = append(emojis, extractUnicode(s[:loc[0]])...)
		emojis = append(emojis, s[loc[0]:loc[1]])

		s = s[loc[1]:]
	}

	return emojis
}

func extractUnicode(s string) []string {
	var out []string

	g := uniseg.NewGraphemes(s)

	for g.Next() {
		cluster := g.Str()

		if isemoji.IsEmoji(cluster) {
			out = append(out, cluster)
		}
	}

	return out
}
