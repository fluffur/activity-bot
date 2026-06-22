package user

import (
	"fmt"
	"strings"
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
	Emojis    string
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
		IsBot:     isBot,
		CreatedAt: now,
	}
}

func (u User) FullName() string {
	return strings.TrimSpace(fmt.Sprintf("%s %s", u.FirstName, u.LastName))
}
