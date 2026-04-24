package model

import (
	"html"

	"github.com/gotd/td/tg"
)

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
	Emojis        Emojis
	CustomEmojiID string
	IsBot         bool
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

func (u User) AsInput() *tg.InputUser {
	return &tg.InputUser{
		UserID: u.ID,
	}
}

type EmojiType string

const (
	EmojiTypeUnicode EmojiType = "unicode"
	EmojiTypeCustom  EmojiType = "custom"
)

type Emoji struct {
	ID   int64     `json:"id,omitempty"`
	Type EmojiType `json:"type"`
	Char string    `json:"char"`
}

type Emojis []Emoji
