package application

import (
	"activity-bot/internal/info/genshin"
)

type Application struct {
	UserID   int64        `json:"user_id"`
	ChatID   int64        `json:"chat_id"`
	Role     genshin.Role `json:"role"`
	Username string       `json:"username"`
}
