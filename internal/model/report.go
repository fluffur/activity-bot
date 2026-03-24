package model

import (
	"time"
)

type MessageReportMember struct {
	ChatMember          ChatMember
	MessagesCount       int
	NormWarn            int
	NormBan             int
	NewbieThresholdDays int
}

type RestMember struct {
	ChatMember ChatMember
	RestUntil  time.Time
	Status     string
	Tag        string
}

type MemberStats struct {
	ChatMember        ChatMember
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
	Status          Status
	Tag             string
	LeftAt          time.Time
}
