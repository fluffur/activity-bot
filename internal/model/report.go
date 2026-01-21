package model

import (
	"time"
)

type WeeklyMessageReportRow struct {
	UserID        int64
	DisplayName   string
	MessagesCount int32
	WeeklyNorm    int32
	NormDone      bool
}

type ExemptUsersRow struct {
	UserID      int64
	DisplayName string
	ExemptUntil time.Time
}
