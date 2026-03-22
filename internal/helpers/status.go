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
