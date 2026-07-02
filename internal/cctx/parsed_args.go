package cctx

import (
	"activity-bot/internal/chatmember"
	"time"
)

type ParsedArgs struct {
	Users     []chatmember.ChatMember
	Numbers   []int64
	Durations []time.Duration
	DateTimes []time.Time
	Texts     []string
}
