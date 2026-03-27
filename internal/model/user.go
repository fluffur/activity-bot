package model

import "html"

const (
	GenderMale   = "male"
	GenderFemale = "female"
)

type User struct {
	ID            int64
	FirstName     string
	LastName      string
	Username      string
	Gender        string
	Emoji         string
	CustomEmojiID string
}

func (u User) DisplayName() string {
	if u.Username != "" {
		return html.EscapeString(u.Username)
	}
	name := u.FirstName
	if u.LastName != "" {
		name += " " + u.LastName
	}
	return html.EscapeString(name)
}
