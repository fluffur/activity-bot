package model

import (
	"time"
)

type MessageReportMember struct {
	User                User
	MessagesCount       int32
	WeeklyNorm          int32
	NormDone            bool
	JoinedAt            time.Time
	NewbieThresholdDays int32
	Role                string
	CustomTitle         string
}

type ExemptMember struct {
	User        User
	ExemptUntil time.Time
	Role        string
	CustomTitle string
}
