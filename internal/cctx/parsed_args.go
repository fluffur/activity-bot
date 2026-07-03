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

func (a ParsedArgs) User() (chatmember.ChatMember, bool) {
	if len(a.Users) == 0 {
		return chatmember.ChatMember{}, false
	}

	return a.Users[0], true
}

func (a ParsedArgs) Number() (int64, bool) {
	if len(a.Numbers) == 0 {
		return 0, false
	}

	return a.Numbers[0], true
}

func (a ParsedArgs) Duration() (time.Duration, bool) {
	if len(a.Durations) != 0 {
		return 0, false
	}

	return a.Durations[0], true
}

func (a ParsedArgs) DateTime() (time.Time, bool) {
	if len(a.DateTimes) == 0 {
		return time.Time{}, false
	}

	return a.DateTimes[0], true
}

func (a ParsedArgs) Text() (string, bool) {
	if len(a.Texts) == 0 {
		return "", false
	}

	return a.Texts[0], true
}

func (a ParsedArgs) Until() (time.Time, bool) {
	if len(a.DateTimes) != 0 {
		return a.DateTimes[0], true
	}

	if len(a.Durations) != 0 {
		return time.Now().Add(a.Durations[0]), true
	}

	return time.Time{}, false
}
