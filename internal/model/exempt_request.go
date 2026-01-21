package model

import (
	"time"
)

type ExemptRequest struct {
	ChatID      int64
	UserID      int64
	RequestedAt time.Time
	ExemptUntil time.Time
	Status      string
	MessageID   int64
}
