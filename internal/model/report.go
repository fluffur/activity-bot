package model

type MessageReportMember struct {
	ChatMember    ChatMember
	Chat          Chat
	MessagesCount int64
}

type MemberStats struct {
	ChatMember        ChatMember
	Chat              Chat
	DayCount          int64
	DayRollingCount   int64
	WeekCount         int64
	WeekRollingCount  int64
	MonthCount        int64
	MonthRollingCount int64
	AllTime           int64
}
