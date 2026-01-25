package common

type AdminService interface {
	IsAdmin(chatID, userID int64) (bool, error)
}
