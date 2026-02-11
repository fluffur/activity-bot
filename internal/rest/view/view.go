package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"time"
)

func FormatRestSet(user model.User, date time.Time, isSelf bool) string {
	if isSelf {
		return fmt.Sprintf("Вы добавлены в рест до %s", helpers.FormatToHumanDate(date))
	}
	return fmt.Sprintf("Пользователь %s добавлен в рест до %s", helpers.Link(user), helpers.FormatToHumanDate(date))
}

func FormatRestRequest(user model.User, date time.Time) string {
	return fmt.Sprintf(
		"Для пользователя %s запрошен рест до %s",
		helpers.Link(user),
		helpers.FormatToHumanDate(date),
	)
}

func FormatRestShow(user model.User, restUntil *time.Time) string {
	if restUntil == nil {
		return fmt.Sprintf("Пользователь %s не находится в ресте", helpers.Link(user))
	}
	return fmt.Sprintf("Пользователь %s находится в ресте до %s", helpers.Link(user), helpers.FormatToHumanDate(*restUntil))
}

func FormatRestEnded(user model.User, isSelf bool) string {
	if isSelf {
		return "Вы успешно удалены из реста"
	}
	return fmt.Sprintf("Пользователь %s успешно удалён из реста", helpers.Link(user))
}

func FormatRestNotInRest(user model.User, isSelf bool) string {
	if isSelf {
		return "Вы не находитесь в ресте"
	}
	return fmt.Sprintf("Пользователь %s не находится в ресте", helpers.Link(user))
}

func FormatRestRequestApproved(user model.User, restUntil time.Time) string {
	return fmt.Sprintf("Запрос одобрен. У %s рест до %s", helpers.Link(user), helpers.FormatToHumanDate(restUntil))
}

func FormatRestRequestRejected(user *model.User) string {
	if user == nil {
		return "Запрос на рест отклонён"
	}
	return fmt.Sprintf("Запрос на рест для %s отклонён", helpers.Link(*user))
}
