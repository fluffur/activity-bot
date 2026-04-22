package marriage

import "activity-bot/internal/model"
import "time"

type Marriage struct {
	ID        int64
	ChatID    int64
	MarriedAt time.Time
	User1     model.ChatMember
	User2     model.ChatMember
}

type RequestStatus string

const (
	RequestStatusPending   RequestStatus = "pending"
	RequestStatusAccepted  RequestStatus = "accepted"
	RequestStatusRejected  RequestStatus = "rejected"
	RequestStatusCancelled RequestStatus = "cancelled"
)

type MarriageRequest struct {
	ID         int64
	ChatID     int64
	FromUserID int64
	ToUserID   int64
	Status     RequestStatus
}
