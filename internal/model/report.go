package model

import (
	"time"
)

type WeeklyMessageReportMember struct {
	User                User
	MessagesCount       int32
	WeeklyNorm          int32
	NormDone            bool
	JoinedAt            time.Time
	NewbieThresholdDays int32
}

type ExemptMember struct {
	User        User
	ExemptUntil time.Time
}
