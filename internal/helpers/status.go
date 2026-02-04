package helpers

func TranslateMemberStatus(status string) string {

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
