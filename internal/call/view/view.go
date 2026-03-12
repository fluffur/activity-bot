package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"math/rand"
	"regexp"
	"strings"
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

func FormatCallChunk(message string, members []model.ChatMember, mentionTypes int32) string {
	var sb strings.Builder
	if message != "" {
		sb.WriteString(message)

		if mentionTypes != 0 {
			sb.WriteString("\n\n")
		}
	}

	separator := " "
	if mentionTypes&MentionTypeName > 0 || mentionTypes&MentionTypeRole > 0 {
		separator = ", "
	}

	for j, m := range members {
		var parts []string
		emptyStr := "​"
		if j == 0 && message == "" {
			emptyStr = "ㅤ"
		}

		emoji := userEmoji(m.User)

		if mentionTypes&MentionTypeEmoji > 0 {
			parts = append(parts, emoji)
		}
		if mentionTypes&MentionTypeName > 0 {
			parts = append(parts, m.User.FirstName)
		}
		if mentionTypes&MentionTypeRole > 0 && m.CustomTitle != "" {
			parts = append(parts, m.CustomTitle)
		}

		if len(parts) == 0 {
			parts = append(parts, emptyStr)
		}

		title := strings.Join(parts, " ")
		if strings.TrimSpace(title) == "" {
			title = emptyStr
		}

		if mentionTypes&MentionTypeEmoji > 0 && hasCustomEmoji(emoji) {
			sb.WriteString(emoji)

			if mentionTypes&(MentionTypeName|MentionTypeRole) > 0 {
				sb.WriteString(" ")
				sb.WriteString(helpers.Mention(m.User.ID, title))
			} else {
				sb.WriteString(" ")
				sb.WriteString(helpers.Mention(m.User.ID, "ㅤ"))
			}
		} else {
			sb.WriteString(helpers.Mention(m.User.ID, title))
		}

		if j < len(members)-1 {
			sb.WriteString(separator)
		}
	}

	return sb.String()
}

func FormatWelcomeCallMessageSet() string {
	return "Новое сообщение для call установлено"
}

func FormatCallOnJoinEnabled() string {
	return "Теперь при инвайте новых участников будет вызываться call"
}

func FormatCallOnJoinDisabled() string {
	return "Теперь при инвайте новых участников не будет вызываться call"
}

func FormatWelcomeCallMessage(message string) string {
	if message == "" {
		return "Сообщение ещё не указано"
	}
	return "Сообщение: " + message
}

func userEmoji(u model.User) string {
	if u.Emoji != "" {
		return u.Emoji
	}

	return callEmojis[rand.Intn(len(callEmojis))]
}

func hasCustomEmoji(s string) bool {
	return strings.Contains(s, "<tg-emoji")
}
