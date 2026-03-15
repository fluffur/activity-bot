package model

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
