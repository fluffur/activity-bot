package model

import (
	"time"
)

type MessageReportMember struct {
	User                User
	MessagesCount       int32
	NormWarn            int32
	NormBan             int32
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

type MemberStats struct {
	User User

	DayCount          int32
	DayRollingCount   int32
	WeekCount         int32
	WeekRollingCount  int32
	MonthCount        int32
	MonthRollingCount int32
	AllTime           int32

	NormBan         int32
	NormWarn        int32
	JoinedAt        time.Time
	RestUntil       *time.Time
	NewbieThreshold int32
	Status          string
	CustomTitle     *string
}
