package application

type Application struct {
	UserID        int64  `json:"user_id"`
	Role          string `json:"role"`
	Username      string `json:"username"`
	ApplicationID int    `json:"application_id"`
}
