package model

import "time"

type Chat struct {
	ID                  int64
	WeeklyNorm          int32
	NewbieThresholdDays int32
	GeminiSystemPrompt  string
}

type ChatMember struct {
	User        User
	ChatID      int64
	ExemptUntil *time.Time
	CustomTitle string
	Role        string
}
