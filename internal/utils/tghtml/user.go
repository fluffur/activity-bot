package tghtml

import "fmt"

func UserMention(userID int64, text string) string {
	return Link(fmt.Sprintf("tg://user?id=%d", userID), text)
}

func Link1(username, text string, userID int64) string {
	if username == "" {
		return Link(fmt.Sprintf("tg://openmessage?user_id=%d", userID), text)
	}

	return Link(fmt.Sprintf("t.me/%s", username), text)
}

func UserLink(username string) string {
	return Link(fmt.Sprintf("t.me/%s", username), "@"+username)
}

func Link(href, content string) string {
	return fmt.Sprintf("<a href=%q>%s</a>", href, content)
}

func Bold(text string) string {
	return "<b>" + text + "</b>"
}
