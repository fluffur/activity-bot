package tghtml

import (
	"activity-bot/internal/emoji"
	"activity-bot/internal/permission"
)

func PatPatEmoji() string {
	return Emoji(emoji.PatPatID, emoji.PatPatFallback)
}

func NewbieEmoji() string {
	return Emoji(emoji.NewbieID, emoji.NewbieFallback)
}

func TotalEmoji() string {
	return Emoji(emoji.TotalID, emoji.TotalFallback)
}

func RestEmoji() string {
	return Emoji(emoji.RestID, emoji.RestFallback)
}

func RestEmoji2() string {
	return Emoji(emoji.Rest2ID, emoji.Rest2Fallback)
}

func SuccessEmoji() string {
	return Emoji(emoji.SuccessID, emoji.SuccessFallback)
}

func DangerEmoji() string {
	return Emoji(emoji.DangerID, emoji.DangerFallback)
}

func StatsEmoji() string {
	return Emoji(emoji.StatsID, emoji.StatsFallback)
}

func StatusEmoji(status permission.Status) string {
	switch status {
	case permission.StatusMember:
		return Emoji(emoji.StatusMemberID, emoji.StatusMemberFallback)
	case permission.StatusModerator:
		return Emoji(emoji.StatusModeratorID, emoji.StatusModeratorFallback)
	case permission.StatusAdmin:
		return Emoji(emoji.StatusAdminID, emoji.StatusAdminFallback)
	case permission.StatusSeniorAdmin:
		return Emoji(emoji.StatusSeniorAdminID, emoji.StatusSeniorAdminFallback)
	case permission.StatusCoOwner:
		return Emoji(emoji.StatusCoOwnerID, emoji.StatusCoOwnerFallback)
	case permission.StatusOwner:
		return Emoji(emoji.StatusOwnerID, emoji.StatusOwnerFallback)
	default:
		return ""
	}
}

func CalendarEmoji() string {
	return Emoji(emoji.CalendarID, emoji.CalendarFallback)
}

func ChartEmoji() string {
	return Emoji(emoji.ChartID, emoji.ChartFallback)
}

func RollingEmoji() string {
	return Emoji(emoji.RollingID, emoji.RollingFallback)
}

func ProfileEmoji() string {
	return Emoji(emoji.ProfileID, emoji.ProfileFallback)
}

func DescEmoji() string {
	return Emoji(emoji.DescID, emoji.DescFallback)
}
