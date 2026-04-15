package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/gotd/td/telegram/message/entity"
)

const (
	MentionTypeNWSP  = 0
	MentionTypeEmoji = 1 << iota
	MentionTypeName
	MentionTypeRole
)

var callEmojis = []string{
	"🔔", "📢", "📣", "⚡️", "✨", "🌟", "🔥", "🌈", "☄️", "🚀",
	"💎", "🧿", "🔮", "🍀", "🌸", "🌺", "🌼", "🌻", "🌿", "🍃",
	"🌍", "🌙", "🔆", "🎵", "🎶", "🎨", "🎭", "🎪", "🎬", "🎤",
	"🏆", "🏅", "🎖", "🎟", "🧘", "🧩", "🪁", "🛰", "⚓️", "🛸",
	"💫", "⭐️", "🌠", "🌌", "🪐", "🌊", "💥", "🎇", "🎆", "🕊",
	"👑", "💖", "💙", "💜", "🤍", "💛", "🧡", "❤️‍🔥", "💗", "💞",
	"🪄", "🎀", "🦋", "🐚", "🌷", "🌹", "🌾", "🍓", "🍒", "🍇",
	"🥂", "🍹", "🧁", "🍩", "🍪", "🌞", "🌤", "⛅️", "🌅", "🌄",
	"🌀", "💠", "🧡", "💚", "🤎", "🖤", "🩵", "🩷", "🪻", "🪷",
}

var tgMentionRegex = regexp.MustCompile(`(?i)<a\s+href="tg://user\?id=\d+">([^<]+)</a>`)
var mentionRegex = regexp.MustCompile(`(?i)(^|[^A-Za-z0-9_])@([a-zA-Z0-9_]{5,32})`)

func ReplaceMentionsWithLinks(input string) string {
	input = tgMentionRegex.ReplaceAllString(input, "<a href=\"tg://openmessage?user_id=$2\">$1</a>")
	var sb strings.Builder
	inTag := false
	var textBuf strings.Builder

	flushText := func() {
		if textBuf.Len() > 0 {
			res := mentionRegex.ReplaceAllString(textBuf.String(), "$1<a href=\"https://t.me/$2\">@$2</a>")
			sb.WriteString(res)
			textBuf.Reset()
		}
	}

	for _, r := range input {
		if r == '<' {
			flushText()
			inTag = true
			sb.WriteRune(r)
		} else if r == '>' && inTag {
			inTag = false
			sb.WriteRune(r)
		} else if inTag {
			sb.WriteRune(r)
		} else {
			textBuf.WriteRune(r)
		}
	}
	flushText()
	return sb.String()
}

func FormatCallChunkBuilder(eb *entity.Builder, message string, members []model.ChatMember, mentionTypes int32) {
	if message != "" {
		eb.Plain(message)

		if mentionTypes != 0 {
			eb.Plain("\n\n")
		}
	}

	separator := " "
	if mentionTypes&MentionTypeName > 0 || mentionTypes&MentionTypeRole > 0 {
		separator = ", "
	}

	for j, m := range members {
		RenderMention(eb, m, mentionTypes)

		if j < len(members)-1 {
			eb.Plain(separator)
		}
	}
	eb.Plain("ㅤ")

}

func FormatWelcomeCallMessage(message string) string {
	if message == "" {
		return "Сообщение ещё не указано"
	}
	return ": " + message
}

func userEmojis(m model.ChatMember) model.Emojis {
	if len(m.Emojis) != 0 {
		return m.Emojis
	}

	if len(m.User.Emojis) != 0 {
		return m.User.Emojis
	}

	return model.Emojis{
		{
			Type: model.EmojiTypeUnicode,
			Char: callEmojis[rand.Intn(len(callEmojis))],
		},
	}
}

func RenderMention(eb *entity.Builder, m model.ChatMember, mentionTypes int32) {
	hasEmoji := mentionTypes&MentionTypeEmoji > 0
	hasName := mentionTypes&MentionTypeName > 0
	hasRole := mentionTypes&MentionTypeRole > 0

	var resultTitle string
	if hasName && hasRole {
		resultTitle = fmt.Sprintf("%s (%s)", m.User.FirstName, m.Tag)
	} else if hasName {
		resultTitle = m.User.FirstName
	} else if hasRole {
		resultTitle = m.Tag
	}

	if !hasEmoji {
		helpers.WriteMention(eb, m.User.ID, resultTitle)
		return
	}

	helpers.MentionEmoji(eb, m.User, userEmojis(m), resultTitle)

}
