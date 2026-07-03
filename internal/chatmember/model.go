package chatmember

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/permission"
	"activity-bot/internal/user"
	"strings"
	"time"
)

type ChatMember struct {
	User            user.User
	Chat            chat.Chat
	RestUntil       time.Time
	RestReason      string
	Tag             string
	Status          permission.Status
	Emojis          string
	JoinedAt        time.Time
	LeftAt          time.Time
	ExcludeFromCall bool
}

func New(u user.User, c chat.Chat, tag string, status permission.Status, now time.Time) ChatMember {
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

func (c ChatMember) Permitted(status permission.Status) bool {
	return c.Status >= status
}

func (c ChatMember) AnyEmoji() string {
	if c.Emojis != "" {
		return c.Emojis
	}

	return c.User.Emojis
}

func (c ChatMember) IsResting(now time.Time) bool {
	return c.RestUntil.After(now)
}

func (c ChatMember) IsNewbie(now time.Time, thresholdDays int32) bool {
	return c.JoinedAt.After(
		now.AddDate(0, 0, -int(thresholdDays)),
	)
}

func (c ChatMember) IsLeft() bool {
	return !c.LeftAt.IsZero()
}

func (c ChatMember) IsMale() bool {
	return c.User.IsMale()
}

func (c ChatMember) IsFemale() bool {
	return c.User.IsFemale()
}

func (c ChatMember) ID() int64 {
	return c.User.ID
}

func (c ChatMember) IsOwner() bool {
	return c.Status == permission.StatusOwner
}
