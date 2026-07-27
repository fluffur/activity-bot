package crocodile

import "time"

type Game struct {
	ChatID       int64     `json:"chat_id"`
	HostID       int64     `json:"host_id"`
	Word         string    `json:"word"`
	StartedAt    time.Time `json:"started_at"`
	SkipCount    int       `json:"skip_count"`
	SkippedWords []string  `json:"skipped_words"`
}

type Word struct {
	ID         int64
	Word       string
	Category   string
	Difficulty int16
	UsedCount  int32
}
