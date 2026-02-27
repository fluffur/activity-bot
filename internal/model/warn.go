package model

import "time"

type Warn struct {
	ID        int64
	User      User
	Moderator User
	Reason    string
	CreatedAt time.Time
	ExpiresAt time.Time
}
