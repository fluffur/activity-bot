package application

type Application struct {
	UserID   int64  `json:"user_id"`
	ChatID   int64  `json:"chat_id"`
	Role     string `json:"role"`
	Username string `json:"username"`
}
