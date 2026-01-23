package model

import (
	"time"
)

type WeeklyMessageReportMember struct {
	User          User
	MessagesCount int32
	WeeklyNorm    int32
	NormDone      bool
}

type ExemptMember struct {
	User        User
	ExemptUntil time.Time
}
