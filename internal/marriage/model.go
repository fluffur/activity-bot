package marriage

import (
	"activity-bot/internal/chatmember"
	"time"
)

type Marriage struct {
	ID        int64
	ChatID    int64
	MarriedAt time.Time
	User1     chatmember.ChatMember
	User2     chatmember.ChatMember
}

type RequestStatus string

const (
	RequestStatusPending   RequestStatus = "pending"
	RequestStatusAccepted  RequestStatus = "accepted"
	RequestStatusRejected  RequestStatus = "rejected"
	RequestStatusCancelled RequestStatus = "cancelled"
)

type Request struct {
	ID         int64
	ChatID     int64
	FromUserID int64
	ToUserID   int64
	Status     RequestStatus
}
