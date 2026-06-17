package tghtml

import "fmt"

func Mention(userID int64, text string) string {
	return Anchor(fmt.Sprintf("tg://user?id=%d", userID), text)
}

func Link(username, text string, userID int64) string {
	if username == "" {
		return Anchor(fmt.Sprintf("tg://openmessage?user_id=%d", userID), text)
	}

	return Anchor(fmt.Sprintf("t.me/%s", username), text)
}

func Anchor(href, content string) string {
	return fmt.Sprintf("<a href=%q>%s</a>", href, content)
}
