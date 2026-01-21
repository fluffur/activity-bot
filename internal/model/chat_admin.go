package model

import "time"

type ChatAdmin struct {
	UserID      int64
	DisplayName string
	CreatedAt   time.Time
}
