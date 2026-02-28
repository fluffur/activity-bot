package model

import "time"

type RestoreRolePayload struct {
	ChatID int64 `json:"chat_id"`
	UserID int64 `json:"user_id"`
}

type RestExpirePayload struct {
	ChatID    int64     `json:"chat_id"`
	UserID    int64     `json:"user_id"`
	RestUntil time.Time `json:"rest_until"`
}
