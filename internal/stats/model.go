package stats

import (
	"activity-bot/internal/chatmember"
)

type ChatStats struct {
	ChatMember    chatmember.ChatMember
	MessagesCount int64
}

type ProfileNorm struct {
	Name     string
	Required int32
	Current  int64
	Passed   bool
}

type ProfileStats struct {
	ChatMember chatmember.ChatMember

	DayCount          int64
	DayRollingCount   int64
	WeekCount         int64
	WeekRollingCount  int64
	MonthCount        int64
	MonthRollingCount int64
	AllTimeCount      int64

	Norms []ProfileNorm
}
