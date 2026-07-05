package tghtml

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/chatmember"
	"activity-bot/internal/i18n"
	"activity-bot/internal/permission"
	"fmt"
	"html"
	"strings"
	"time"
)

func Escape(text string) string {
	return html.EscapeString(text)
}

func UserMention(userID int64, text string) string {
	return Link(fmt.Sprintf("tg://user?id=%d", userID), Escape(text))
}

func MemberLink(
	loc *i18n.Localizer,
	ch chat.Chat,
	member chatmember.ChatMember,
) string {
	return UserLink(
		member.User.Username,
		member.Display(
			loc.T(i18n.User.Unknown, nil),
			ch.EmojisEnabled,
		),
		member.ID(),
	)
}

func MemberMention(
	loc *i18n.Localizer,
	ch chat.Chat,
	member chatmember.ChatMember,
) string {
	return UserMention(member.User.ID, member.Display(loc.T(i18n.User.Unknown, nil), ch.EmojisEnabled))
}

func UserLink(username, text string, userID int64) string {
	if username == "" {
		return Link(fmt.Sprintf("tg://openmessage?user_id=%d", userID), Escape(text))
	}

	return Link(fmt.Sprintf("https://t.me/%s", username), text)
}

func Link(href, content string) string {
	return fmt.Sprintf("<a href=%q>%s</a>", href, content)
}

func StartGroupLink(username string) string {
	return fmt.Sprintf("t.me/%s?startgroup=true", username)
}

func Bold(text string) string {
	return "<b>" + text + "</b>"
}

func Italic(text string) string {
	return "<i>" + Escape(text) + "</i>"
}

func Code(text string) string {
	return "<code>" + Escape(text) + "</code>"
}

func Pre(text string) string {
	return "<pre>" + Escape(text) + "</pre>"
}

func Blockquote(text string) string {
	return "<blockquote>" + text + "</blockquote>"
}

func ExpandableBlockquote(text string) string {
	text = strings.TrimSpace(text)

	return fmt.Sprintf(
		"<blockquote expandable>%s</blockquote>",
		text,
	)
}

func DefaultDateTime(t time.Time) string {
	return DateTime(t, "wdt", t.Format("02.01.2006"))
}

func RelativeDateTime(from, to time.Time) string {
	if from.After(to) {
		to, from = from, to
	}

	years := to.Year() - from.Year()
	months := int(to.Month()) - int(from.Month())
	days := to.Day() - from.Day()

	if days < 0 {
		months--
	}

	if months < 0 {
		years--

		months += 12
	}

	var fallback string

	switch {
	case years > 0:
		fallback = plural(years, "год", "года", "лет")
	case months > 0:
		fallback = plural(months, "месяц", "месяца", "месяцев")
	default:
		d := int(to.Sub(from).Hours() / 24)
		if d <= 0 {
			d = 1
		}

		fallback = plural(d, "день", "дня", "дней")
	}

	return DateTime(from, "r", fallback)
}

func plural(n int, one, few, many string) string {
	n10 := n % 10
	n100 := n % 100

	switch {
	case n10 == 1 && n100 != 11:
		return fmt.Sprintf("%d %s", n, one)
	case n10 >= 2 && n10 <= 4 && (n100 < 12 || n100 > 14):
		return fmt.Sprintf("%d %s", n, few)
	default:
		return fmt.Sprintf("%d %s", n, many)
	}
}

func DateTime(t time.Time, format, fallback string) string {
	if abs := time.Since(t); abs < 0 {
		abs = -abs
		if abs > 3*365*24*time.Hour {
			return fallback
		}
	} else if abs > 3*365*24*time.Hour {
		return fallback
	}

	return fmt.Sprintf(
		"<tg-time unix=\"%d\" format=%q>%s</tg-time>",
		t.Unix(),
		format,
		fallback,
	)
}

func Emoji(customEmojiID int, fallbackEmoji string) string {
	if strings.TrimSpace(fallbackEmoji) != "" {
		return fallbackEmoji
	}

	return fmt.Sprintf("<tg-emoji id=%d>%s</tg-emoji>", customEmojiID, fallbackEmoji)
}

func Status(loc *i18n.Localizer, s permission.Status) string {
	return loc.T(s.TranslationKey(), nil)
}
