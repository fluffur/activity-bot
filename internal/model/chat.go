package model

import (
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/gotd/td/tg"
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
	MaxWarns            int32
}

func (c Chat) AsInputChannel() *tg.InputChannel {
	return &tg.InputChannel{
		ChannelID: c.ID,
	}
}

type Status int16

const (
	StatusMember Status = iota
	StatusModerator
	StatusAdmin
	StatusSeniorAdmin
	StatusCoOwner
	StatusOwner
)

func (s Status) String() string {
	switch s {
	case StatusMember:
		return "участник"
	case StatusModerator:
		return "модератор"
	case StatusAdmin:
		return "младший администратор"
	case StatusSeniorAdmin:
		return "старший администратор"
	case StatusCoOwner:
		return "совладелец"
	case StatusOwner:
		return "владелец"
	}
	return "неизвестно"
}

func (s Status) Plural() string {
	switch s {
	case StatusMember:
		return "участники"
	case StatusModerator:
		return "модераторы"
	case StatusAdmin:
		return "младшие администраторы"
	case StatusSeniorAdmin:
		return "старшие администраторы"
	case StatusCoOwner:
		return "совладельцы"
	case StatusOwner:
		return "владельцы"
	}
	return "неизвестно кто"
}

func (s Status) PluralTitle() string {
	return cases.Title(language.Und, cases.NoLower).String(s.Plural())
}
func (s Status) Title() string {
	return cases.Title(language.Und, cases.NoLower).String(s.String())
}

type ChatMember struct {
	User            User
	ChatID          int64
	RestUntil       time.Time
	RestReason      string
	Tag             string
	Status          Status
	Emoji           string
	Emojis          Emojis
	JoinedAt        time.Time
	LeftAt          time.Time
	ExcludeFromCall bool
}

func (cm ChatMember) CanModerate(c ChatMember) bool {
	return cm.Status > c.Status
}

func (cm ChatMember) StatusGranted(s Status) bool {
	return cm.Status >= s
}

func (cm ChatMember) IsRestActive(now time.Time) bool {
	return !cm.RestUntil.IsZero() && cm.RestUntil.After(now)
}

func (cm ChatMember) AsInputChannel() *tg.InputChannel {
	return &tg.InputChannel{
		ChannelID: cm.ChatID,
	}
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
