package model

import (
	"time"
)

type WeeklyMessageReportMember struct {
	UserID        int64
	FullName      string
	MessagesCount int32
	WeeklyNorm    int32
	NormDone      bool
	Username      *string
}

type ExemptMember struct {
	UserID      int64
	FullName    string
	ExemptUntil time.Time
	Username    *string
}
