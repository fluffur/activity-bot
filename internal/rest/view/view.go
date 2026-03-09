package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"time"
)

func FormatRestSet(user model.User, date time.Time, reason string) string {
	text := fmt.Sprintf("Участник %s %s в рест до %s",
		helpers.UserLink(user),
		helpers.Gendered(user.Gender, "добавлен", "добавлена", "добавлен(а)"),
		helpers.FormatToHumanDateTime(date),
	)
	if reason != "" {
		text += fmt.Sprintf("\n\nПричина: %s", reason)
	}
	return text
}

func FormatRestRequest(user model.User, date time.Time, reason string) string {
	text := fmt.Sprintf(
		"Для участника %s запрошен рест до %s",
		helpers.UserLink(user),
		helpers.FormatToHumanDateTime(date),
	)
	if reason != "" {
		text += fmt.Sprintf("\n\nПричина: %s", reason)
	}
	return text
}

func FormatRestShow(m model.ChatMember) string {
	if m.RestUntil == nil {
		return fmt.Sprintf("Участник %s не находится в ресте", helpers.UserLink(m.User))
	}
	text := fmt.Sprintf("Участник %s находится в ресте до %s", helpers.UserLink(m.User), helpers.FormatToHumanDateTime(*m.RestUntil))
	if m.RestReason != "" {
		text += fmt.Sprintf("\n\nПричина: %s", m.RestReason)
	}
	return text
}

func FormatRestEnded(user model.User, isSelf bool) string {
	if isSelf {
		return "Вы успешно удалены из реста"
	}
	return fmt.Sprintf("Участник %s успешно %s из реста",
		helpers.UserLink(user),
		helpers.Gendered(user.Gender, "удалён", "удалена", "удалён(а)"),
	)
}

func FormatRestNotInRest(user model.User, isSelf bool) string {
	if isSelf {
		return "Вы не находитесь в ресте"
	}
	return fmt.Sprintf("Пользователь %s не находится в ресте", helpers.UserLink(user))
}

func FormatRestRequestApproved(user model.User, restUntil time.Time) string {
	return fmt.Sprintf("Запрос одобрен. У %s рест до %s", helpers.UserLink(user), helpers.FormatToHumanDateTime(restUntil))
}

func FormatRestRequestRejected(user *model.User) string {
	if user == nil {
		return "Запрос на рест отклонён"
	}
	return fmt.Sprintf("Запрос на рест для %s отклонён", helpers.UserLink(*user))
}
