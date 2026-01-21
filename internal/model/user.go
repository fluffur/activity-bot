package model

type User struct {
	ID        int64
	FirstName string
	LastName  string
	Username  *string
}

func NewUser(id int64) User {
	return User{
		ID: id,
	}
}
