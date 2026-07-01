package chatmember

import (
	"activity-bot/internal/chat"
	"activity-bot/internal/i18n"
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

func (c ChatMember) Permitted(status Status) bool {
	return c.Status >= status
}

func (c ChatMember) AnyEmoji() string {
	if c.Emojis != "" {
		return c.Emojis
	}

	return c.User.Emojis
}

func (s Status) TranslationKey() i18n.MessageID {
	switch s {
	case StatusMember:
		return i18n.Status.Member
	case StatusModerator:
		return i18n.Status.Moderator
	case StatusAdmin:
		return i18n.Status.JuniorAdmin
	case StatusSeniorAdmin:
		return i18n.Status.SeniorAdmin
	case StatusCoOwner:
		return i18n.Status.Coowner
	case StatusOwner:
		return i18n.Status.Owner
	default:
		return i18n.Status.Member
	}
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

func (s Status) Emoji() string {
	switch s {
	case StatusMember:
		return "0️⃣"
	case StatusModerator:
		return "1️⃣"
	case StatusAdmin:
		return "2️⃣"
	case StatusSeniorAdmin:
		return "3️⃣"
	case StatusCoOwner:
		return "4️⃣"
	case StatusOwner:
		return "5️⃣"
	default:
		return "?"
	}
}

func (c ChatMember) IsMale() bool {
	return c.User.IsMale()
}

func (c ChatMember) IsFemale() bool {
	return c.User.IsFemale()
}
