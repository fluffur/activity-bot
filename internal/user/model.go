package user

import (
	"activity-bot/internal/emoji"
	"time"
)

type Gender string

const (
	GenderMale    Gender = "male"
	GenderFemale  Gender = "female"
	GenderUnknown Gender = "unknown"
)

type User struct {
	ID        int64
	FirstName string
	LastName  string
	Username  string
	Gender    Gender
	Emojis    emoji.Emojis
	IsBot     bool
	CreatedAt time.Time
}

func New(id int64, firstName, lastName, username string, gender Gender, isBot bool, now time.Time) User {
	return User{
		ID:        id,
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
		Gender:    gender,
		Emojis:    emoji.Emojis{},
		IsBot:     isBot,
		CreatedAt: now,
	}
}
