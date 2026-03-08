package helpers

import (
	"activity-bot/internal/model"
	"fmt"
	"html"
	"strings"
)

func Link(u model.User) string {
	if u.Username != nil && *u.Username != "" {
		return fmt.Sprintf(`<a href="https://t.me/%s">%s</a>`, *u.Username, html.EscapeString(u.FirstName))
	}

	return fmt.Sprintf(`<a href="tg://openmessage?user_id=%d">%s</a>`, u.ID, html.EscapeString(u.FirstName))
}

func LinkWithContent(u model.User, content string) string {
	if u.Username != nil && *u.Username != "" {
		return fmt.Sprintf(`<a href="https://t.me/%s">%s</a>`, *u.Username, html.EscapeString(content))
	}

	return fmt.Sprintf(`<a href="tg://openmessage?user_id=%d">%s</a>`, u.ID, html.EscapeString(content))
}

func Mention(id int64, value string) string {
	if value == "" {
		value = "?"
	}
	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, id, html.EscapeString(value))
}

func TelegramMessageLink(chatID int64, messageID int64, username string) string {
	if username != "" {
		return fmt.Sprintf("https://t.me/%s/%d", username, messageID)
	}
	if chatID < 0 {
		id := strings.TrimPrefix(fmt.Sprint(chatID), "-100")
		return fmt.Sprintf("https://t.me/c/%s/%d", id, messageID)
	}
	return ""
}
