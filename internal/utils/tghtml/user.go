package tghtml

import (
	"fmt"
	"html" // Важно для экранирования текста
)

// Escape экранирует спецсимволы, чтобы Telegram не ловил ENTITY_BOUNDS_INVALID
func Escape(text string) string {
	return html.EscapeString(text)
}

func UserMention(userID int64, text string) string {
	return Link(fmt.Sprintf("tg://user?id=%d", userID), Escape(text))
}

func Link1(username, text string, userID int64) string {
	if username == "" {
		return Link(fmt.Sprintf("tg://openmessage?user_id=%d", userID), Escape(text))
	}
	return Link(fmt.Sprintf("t.me/%s", username), Escape(text))
}

func UserLink(username string) string {
	return Link(fmt.Sprintf("t.me/%s", username), "@"+Escape(username))
}

func Link(href, content string) string {
	return fmt.Sprintf("<a href=%q>%s</a>", href, content)
}

func StartGroupLink(username string) string {
	return fmt.Sprintf("t.me/%s?startgroup=true", username)
}

func Bold(text string) string {
	return "<b>" + Escape(text) + "</b>"
}

func Italic(text string) string {
	return "<i>" + Escape(text) + "</i>"
}

// Code теперь возвращает <code>, который корректно работает инлайн
func Code(text string) string {
	return "<code>" + Escape(text) + "</code>"
}

// Pre для полноценных многострочных блоков кода (если понадобятся)
func Pre(text string) string {
	return "<pre>" + Escape(text) + "</pre>"
}
