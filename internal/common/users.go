package common

import "activity-bot/internal/model"

type UserService interface {
	GetUserByUsername(username string) (model.User, error)
	EnsureUserExists(id int64, username, firstName, lastName string) (model.User, error)
}
