package model

const (
	GenderMale    = "male"
	GenderFemale  = "female"
	GenderUnknown = "unknown"
)

type User struct {
	ID        int64
	FirstName string
	LastName  string
	Username  *string
	Gender    string
}
