package model

type RestoreRolePayload struct {
	ChatID int64 `json:"chat_id"`
	UserID int64 `json:"user_id"`
}
