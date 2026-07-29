package application

import (
	"activity-bot/internal/info/genshin"
)

type State string

const (
	AppStateIdle        State = ""
	AppStateAwaitRole   State = "await_role"
	AppStateConfirmRole State = "confirm_role"
	AppStatePending     State = "pending"
)

type AppStateData struct {
	Role genshin.Role `json:"role"`
}

type RejectState string
type RejectStateData struct {
	UserID int64 `json:"user_id"`
	ChatID int64 `json:"chat_id"`
}

const (
	RejectStateIdle               RejectState = ""
	RejectStateAwaitRejectMessage RejectState = "await_reject_message"
)
