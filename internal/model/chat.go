package model

import "time"

type Chat struct {
	ID                  int64
	WeeklyNorm          int32
	NewbieThresholdDays int32
	AISystemPrompt      string
	MaxLadder           int32
	WelcomeCallMessage  string
	CallOnJoin          bool
	WeekStartDay        int16
}

type ChatMember struct {
	User        User
	ChatID      int64
	RestUntil   *time.Time
	CustomTitle string
	Status      string
}

type InactiveMember struct {
	Member       ChatMember
	LastActivity *time.Time
}
