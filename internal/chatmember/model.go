package chatmember

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/emoji"
	"activity-bot/internal/user"
	"time"
)

type Status int16

const (
	StatusMember Status = iota
	StatusModerator
	StatusAdmin
	StatusSeniorAdmin
	StatusCoOwner
	StatusOwner
)

type ChatMember struct {
	User            user.User
	Chat            chat.Chat
	RestUntil       time.Time
	RestReason      string
	Tag             string
	Status          Status
	Emojis          emoji.Emojis
	JoinedAt        time.Time
	LeftAt          time.Time
	ExcludeFromCall bool
}

func New(u user.User, c chat.Chat, tag string, status Status, now time.Time) ChatMember {
	return ChatMember{
		User:     u,
		Chat:     c,
		Tag:      tag,
		Status:   status,
		Emojis:   emoji.Emojis{},
		JoinedAt: now,
	}
}
