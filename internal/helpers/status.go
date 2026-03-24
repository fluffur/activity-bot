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
		CustomEmoji(5458853959487724162, "0️⃣"),
		CustomEmoji(5456218692108950712, "1️⃣"),
		CustomEmoji(5458791343159516783, "2️⃣"),
		CustomEmoji(5458443480873312556, "3️⃣"),
		CustomEmoji(5458747431413882769, "4️⃣"),
		CustomEmoji(5458892313545677652, "5️⃣"),
	}
	if status < 0 || status > 5 {
		return ""
	}
	return customEmojis[status]
}
