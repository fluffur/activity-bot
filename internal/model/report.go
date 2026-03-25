package model

type ChatMemberMessageCount struct {
	ChatMember   ChatMember
	Chat         Chat
	MessageCount int64
}

type ChatMemberStats struct {
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
