package helpers

import (
	"activity-bot/internal/model"
	"time"
)

func TranslateMemberStatus(status model.Status, leftAt time.Time) string {
	if !leftAt.IsZero() {
		return "Не в чате"
	}
	return TranslateMemberStatusNoLeft(status)
}

func TranslateMemberStatusNoLeft(status model.Status) string {
	switch status {
	case model.StatusMember:
		return "Участник"
	case model.StatusModerator:
		return "Модератор"
	case model.StatusAdmin:
		return "Администратор"
	case model.StatusSeniorAdmin:
		return "Старший Администратор"
	case model.StatusCoOwner:
		return "Совладелец"
	case model.StatusOwner:
		return "Владелец"
	}
	return "Неизвестно"
}

func StatusEmoji(status model.Status) string {
	customEmojis := []string{
		CustomEmoji(5368832132158337100, "0️⃣"),
		CustomEmoji(5366311523226496082, "1️⃣"),
		CustomEmoji(5366536356174510823, "2️⃣"),
		CustomEmoji(5366525906519076469, "3️⃣"),
		CustomEmoji(5368367657215078446, "4️⃣"),
		CustomEmoji(5368552151830244939, "5️⃣"),
	}
	if status < 0 || status > 5 {
		return ""
	}
	return customEmojis[status]
}
