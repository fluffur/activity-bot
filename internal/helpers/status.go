package helpers

import "time"

func TranslateMemberStatus(status string, leftAt *time.Time) string {
	if leftAt != nil {
		return "Не в чате"
	}
	switch status {
	case "member":
		return "Участник"
	case "administrator":
		return "Администратор"
	case "creator":
		return "Владелец"

	}
	return "Неизвестно"
}
