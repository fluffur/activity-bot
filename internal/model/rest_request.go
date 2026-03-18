package model

import (
	"time"
)

type RestRequest struct {
	ID          int64
	ChatID      int64
	UserID      int64
	RequestedAt time.Time
	RestUntil   time.Time
	UpdatedAt   time.Time
	Status      string
	MessageID   int64
	Reason      string
}

type ApprovedRestRequest struct {
	RestRequest
	ChatMember ChatMember
}
