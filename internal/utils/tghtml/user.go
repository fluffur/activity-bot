package tghtml

import (
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

func DateTime(time time.Time, format, fallback string) string {
	return fmt.Sprintf("<tg-time unix=\"%d\" format=%q>%s</tg-time>", time.Unix(), format, fallback)
}
