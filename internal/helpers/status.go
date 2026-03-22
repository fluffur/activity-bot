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
		CustomEmoji(5224390118546101411, "0️⃣"),
		CustomEmoji(5224473711494581672, "1️⃣"),
		CustomEmoji(5224251017440285983, "2️⃣"),
		CustomEmoji(5224414625629492488, "3️⃣"),
		CustomEmoji(5224667831131460953, "4️⃣"),
		CustomEmoji(5224483718768384409, "5️⃣"),
	}
	if status < 0 || status > 5 {
		return ""
	}
	return customEmojis[status]
}
