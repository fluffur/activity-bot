package model

import "time"

type Chat struct {
	ID         int64
	WeeklyNorm int32
}

type ChatMember struct {
	ChatID      int64
	UserID      int64
	ExemptUntil *time.Time
}

func NewChat(id int64, weeklyNorm int32) Chat {
	return Chat{
		ID:         id,
		WeeklyNorm: weeklyNorm,
	}
}
