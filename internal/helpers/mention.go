package helpers

import (
	"activity-bot/internal/model"
	"fmt"
	"html"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func Link(u model.User) string {
	if u.Username != nil {
		return fmt.Sprintf(`<a href="https://t.me/%s">%s</a>`, *u.Username, html.EscapeString(u.FirstName))
	}

	return fmt.Sprintf(`<a href="tg://openmessage?user_id=%d">%s</a>`, u.ID, html.EscapeString(u.FirstName))
}

func Mention(u model.User, value string) string {
	return fmt.Sprintf(`<a href="tg://user?id=%d">%s</a>`, u.ID, value)
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
