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
	Tag                 string
	JoinedAt            time.Time
}

type RestMember struct {
	User      User
	RestUntil time.Time
	Status    string
	Tag       string
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
	Status          string
	Tag             string
	LeftAt          time.Time
}
