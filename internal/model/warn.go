package model

import "time"

type Warn struct {
	ID         int64
	ChatMember ChatMember
	Moderator  ChatMember
	Reason     string
	CreatedAt  time.Time
	ExpiresAt  time.Time
}
