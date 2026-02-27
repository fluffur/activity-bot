package helpers

import "time"

func TranslateMemberStatus(status string, leftAt *time.Time) string {
	if leftAt != nil {
		return "не в чате"
	}
	switch status {
	case "member":
		return "участник"
	case "administrator":
		return "администратор"
	case "creator":
		return "владелец"

	}
	return "неизвестно"
}
