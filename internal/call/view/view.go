package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"math/rand"
	"regexp"
	"strconv"
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

		emoji := userEmoji(m)

		if mentionTypes&MentionTypeEmoji > 0 && !hasCustomEmoji(emoji) {
			parts = append(parts, emoji)
		}
		if mentionTypes&MentionTypeName > 0 {
			parts = append(parts, m.User.DisplayName())
		}
		if mentionTypes&MentionTypeRole > 0 && m.Tag != "" {
			parts = append(parts, m.Tag)
		}

		if len(parts) == 0 && !(mentionTypes&MentionTypeEmoji > 0 && hasCustomEmoji(emoji)) {
			parts = append(parts, emptyStr)
		}

		title := strings.Join(parts, " ")
		if strings.TrimSpace(title) == "" {
			title = emptyStr
		}

		if mentionTypes&MentionTypeEmoji > 0 && hasCustomEmoji(emoji) {
			sb.WriteString(emoji)
			sb.WriteString(" ")
			if title != "" {
				sb.WriteString(helpers.Mention(m.User.ID, title))
			} else {
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

func FormatCallChunkBuilder(eb *entity.Builder, message string, members []model.ChatMember, mentionTypes int32) {
	if message != "" {
		eb.WriteString(message)

		if mentionTypes != 0 {
			eb.WriteString("\n\n")
		}
	}

	separator := " "
	if mentionTypes&MentionTypeName > 0 || mentionTypes&MentionTypeRole > 0 {
		separator = ", "
	}

	for j, m := range members {
		if mentionTypes&MentionTypeEmoji > 0 {
			emoji := userEmoji(m)
			if hasCustomEmoji(emoji) {
				re := regexp.MustCompile(`emoji-id="(\d+)">([^<]+)`)
				matches := re.FindStringSubmatch(emoji)
				if len(matches) > 2 {
					id, _ := strconv.ParseInt(matches[1], 10, 64)
					eb.CustomEmoji(matches[2], id)
					eb.WriteString(" ")
				}
			} else {
				eb.WriteString(emoji)
				eb.WriteString(" ")
			}
		}

		var parts []string
		if mentionTypes&MentionTypeName > 0 {
			parts = append(parts, m.User.DisplayName())
		}
		if mentionTypes&MentionTypeRole > 0 && m.Tag != "" {
			parts = append(parts, m.Tag)
		}

		title := strings.Join(parts, " ")
		if title == "" {
			title = "​"
		}

		helpers.WriteMention(eb, m.User.ID, title)

		if j < len(members)-1 {
			eb.WriteString(separator)
		}
	}
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

func userEmoji(m model.ChatMember) string {
	if m.Emoji != "" {
		return m.Emoji
	}
	if m.User.Emoji != "" {
		return m.User.Emoji
	}

	return callEmojis[rand.Intn(len(callEmojis))]
}

func hasCustomEmoji(s string) bool {
	return strings.Contains(s, "<tg-emoji")
}
