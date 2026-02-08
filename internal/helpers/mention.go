package helpers

import (
	"activity-bot/internal/model"
	"fmt"
	"html"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func Link(u model.User) string {
	if u.Username != nil || *u.Username != "" {
		return fmt.Sprintf(`<a href="https://t.me/%s">%s</a>`, *u.Username, html.EscapeString(u.FirstName))
	}

	return fmt.Sprintf(`<a href="tg://openmessage?user_id=%d">%s</a>`, u.ID, html.EscapeString(u.FirstName))
}

func LinkWithContent(u model.User, content string) string {
	if u.Username != nil || *u.Username != "" {
		return fmt.Sprintf(`<a href="https://t.me/%s">%s</a>`, *u.Username, html.EscapeString(content))
	}

	return fmt.Sprintf(`<a href="tg://openmessage?user_id=%d">%s</a>`, u.ID, html.EscapeString(content))
}

func Mention(id int64, value string) string {
	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, id, value)
}

func MapUser(f *gotgbot.User) model.User {
	var username *string
	if f.Username != "" {
		username = &f.Username
	}

	return model.User{
		ID:        f.Id,
		FirstName: f.FirstName,
		LastName:  f.LastName,
		Username:  username,
	}
}
