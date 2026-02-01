package model

import (
	"time"
)

type MessageReportMember struct {
	User                User
	MessagesCount       int32
	WeeklyNorm          int32
	NormDone            bool
	NewbieThresholdDays int32
	Status              string
	CustomTitle         string
	JoinedAt            time.Time
}

type RestMember struct {
	User        User
	RestUntil   time.Time
	Status      string
	CustomTitle string
}
