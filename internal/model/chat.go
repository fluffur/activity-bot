package model

import (
	"time"
)

type Chat struct {
	ID                  int64
	Title               string
	NormWarn            int32
	NormBan             int32
	NewbieThresholdDays int32
	AISystemPrompt      string
	MaxLadder           int32
	WelcomeCallMessage  string
	CallOnJoin          bool
	WeekStartDay        int16
	CommandPrefix       string
	AllowPrefixless     bool
	MentionsPerMessage  int32
	MentionTypes        int32
	TagsEnabled         bool
	WeekStartTime       string
}

type ChatMember struct {
	User        User
	ChatID      int64
	RestUntil   time.Time
	RestReason  string
	CustomTitle string
	Status      string
}

func (cm ChatMember) IsRestActive(now time.Time) bool {
	return !cm.RestUntil.IsZero() && cm.RestUntil.After(now)
}

type ChatWithoutNorm struct {
	ID        int64
	Title     string
	NormBan   int32
	NormWarn  int32
	WeekCount int64
}

type InactiveMember struct {
	Member       ChatMember
	LastActivity time.Time
}
