package helpers

import "time"

func TranslateMemberStatus(status int16, leftAt time.Time) string {
	if !leftAt.IsZero() {
		return "Не в чате"
	}
	switch status {
	case 0:
		return "Участник"
	case 1:
		return "Младший модератор"
	case 2:
		return "Старший модератор"
	case 3:
		return "Администратор"
	case 4:
		return "Совладелец"
	case 5:
		return "Владелец"

	}
	return "Неизвестно"
}

func TranslateMemberStatusNoLeft(status int16) string {
	switch status {
	case 0:
		return "Участник"
	case 1:
		return "Младший модератор"
	case 2:
		return "Старший модератор"
	case 3:
		return "Администратор"
	case 4:
		return "Совладелец"
	case 5:
		return "Владелец"

	}
	return "Неизвестно"
}

func StatusEmoji(status int16) string {
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
