package permission

import (
	"activity-bot/internal/emoji"
	"activity-bot/internal/i18n"
	"activity-bot/internal/utils/tghtml"
)

type Status int16

const (
	StatusMember Status = iota
	StatusModerator
	StatusAdmin
	StatusSeniorAdmin
	StatusCoOwner
	StatusOwner
	StatusDisabled
)

const (
	StatusMin = StatusMember
	StatusMax = StatusOwner
)

func (s Status) IsValid() bool {
	return s >= StatusMember && s <= StatusOwner
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
	case StatusDisabled:
		return i18n.Status.Disabled
	default:
		return i18n.Status.Member
	}
}

func (s Status) Emoji() string {
	switch s {
	case StatusMember:
		return tghtml.Emoji(emoji.StatusMemberID, emoji.StatusMemberFallback)
	case StatusModerator:
		return tghtml.Emoji(emoji.StatusModeratorID, emoji.StatusModeratorFallback)
	case StatusAdmin:
		return tghtml.Emoji(emoji.StatusAdminID, emoji.StatusAdminFallback)
	case StatusSeniorAdmin:
		return tghtml.Emoji(emoji.StatusSeniorAdminID, emoji.StatusSeniorAdminFallback)
	case StatusCoOwner:
		return tghtml.Emoji(emoji.StatusCoOwnerID, emoji.StatusCoOwnerFallback)
	case StatusOwner:
		return tghtml.Emoji(emoji.StatusOwnerID, emoji.StatusOwnerFallback)
	default:
		return ""
	}
}

func (s Status) IsDisabled() bool {
	return s == StatusDisabled
}

func (s Status) IsMember() bool {
	return s == StatusMember
}
