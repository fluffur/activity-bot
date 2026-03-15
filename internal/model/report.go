package model

import (
	"time"
)

type MessageReportMember struct {
	User                User
	MessagesCount       int
	NormWarn            int
	NormBan             int
	NewbieThresholdDays int
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

	DayCount          int
	DayRollingCount   int
	WeekCount         int
	WeekRollingCount  int
	MonthCount        int
	MonthRollingCount int
	AllTime           int

	NormBan         int
	NormWarn        int
	JoinedAt        time.Time
	RestUntil       time.Time
	NewbieThreshold int
	Status          string
	CustomTitle     string
	LeftAt          time.Time
}
