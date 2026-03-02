package model

import (
	"time"
)

type RestRequest struct {
	ChatID      int64
	UserID      int64
	RequestedAt time.Time
	RestUntil   time.Time
	Status      string
	MessageID   int64
	Reason      string
}
