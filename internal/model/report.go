package model

import (
	"time"
)

type WeeklyMessageReportMember struct {
	UserID        int64
	DisplayName   string
	MessagesCount int32
	WeeklyNorm    int32
	NormDone      bool
}

type ExemptMember struct {
	UserID      int64
	DisplayName string
	ExemptUntil time.Time
}
