package common

type AdminService interface {
	IsAdmin(chatID, userID int64) (bool, error)
	IsCreator(chatID, userID int64) (bool, error)
	GetRole(chatID, userID int64) (string, error)
}
