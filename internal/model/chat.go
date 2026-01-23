package model

import "time"

type Chat struct {
	ID         int64
	WeeklyNorm int32
}

type ChatMember struct {
	User        User
	ChatID      int64
	ExemptUntil *time.Time
	CustomTitle string
}
