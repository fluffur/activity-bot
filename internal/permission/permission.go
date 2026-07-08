package permission

import "activity-bot/internal/i18n"

type Status int16

const (
	StatusMember Status = iota
	StatusModerator
	StatusAdmin
	StatusSeniorAdmin
	StatusCoOwner
	StatusOwner
)

const (
	StatusMin = StatusMember
	StatusMax = StatusOwner
)

func IsValidStatus(s int64) bool {
	return s >= int64(StatusMember) && s <= int64(StatusOwner)
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
