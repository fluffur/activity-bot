package chatmember

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/user"
	"strings"
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
	Emojis          string
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
		JoinedAt: now,
	}
}

func (c ChatMember) Display(unknown string, emojis bool) string {
	name := strings.TrimSpace(c.User.FirstName + " " + c.User.LastName)

	if c.Tag != "" {
		name = c.Tag
	}

	if name == "" {
		return unknown
	}

	if emojis && c.Emojis != "" {
		return c.Emojis + " " + name
	}

	if emojis && c.User.Emojis != "" {
		return c.User.Emojis + " " + name
	}

	return name
}
