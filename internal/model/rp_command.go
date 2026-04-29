package model

import "time"

type RPCommand struct {
	ChatID    int64
	Trigger   string
	Template  string
	Emoji     Emojis
	CreatedBy int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
