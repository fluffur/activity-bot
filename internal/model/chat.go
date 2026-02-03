package model

import "time"

type Chat struct {
	ID                  int64
	WeeklyNorm          int32
	NewbieThresholdDays int32
	GeminiSystemPrompt  string
	MaxLadder           int32
	WelcomeCallMessage  string
	CallOnJoin          bool
}

type ChatMember struct {
	User        User
	ChatID      int64
	RestUntil   *time.Time
	CustomTitle string
	Status      string
}
