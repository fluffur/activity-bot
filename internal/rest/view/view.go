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
		helpers.Gendered(user.Gender, "добавлен", "добавлена"),
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
	if m.RestUntil.IsZero() {
		return fmt.Sprintf("Участник %s не находится в ресте", helpers.UserLink(m.User))
	}
	message := "Участник %s находится в ресте до %s"
	if m.RestUntil.Before(time.Now()) {
		message = "Рест участника %s был завершен %s"
	}
	text := fmt.Sprintf(message, helpers.UserLink(m.User), helpers.FormatToHumanDateTime(m.RestUntil))
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
		helpers.Gendered(user.Gender, "удалён", "удалена"),
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
func FormatRestRequests(requests []model.ApprovedRestRequest) string {
	if len(requests) == 0 {
		return "Список рестов пуст"
	}

	var approvedText, pendingText, rejectedText string

	for i, r := range requests {
		reasonPart := ""
		if r.Reason != "" {
			reasonPart = fmt.Sprintf(" (%s)", r.Reason)
		}

		// Показываем дату обновления, если есть
		timePart := ""
		if !r.UpdatedAt.IsZero() {
			timePart = fmt.Sprintf(" [обновлено: %s]", helpers.FormatToHumanDateTime(r.UpdatedAt))
		} else {
			// иначе показываем дату запроса
			timePart = fmt.Sprintf(" [запрошено: %s]", helpers.FormatToHumanDateTime(r.RequestedAt))
		}

		line := fmt.Sprintf("%d. до %s%s%s\n", i+1, helpers.FormatToHumanDateTime(r.RestUntil), reasonPart, timePart)

		switch r.Status {
		case "approved":
			approvedText += line
		case "pending":
			pendingText += line
		case "rejected":
			rejectedText += line
		default:
			pendingText += line
		}
	}

	text := "<b>Список рестов:</b>\n"

	if approvedText != "" {
		text += "\n<b>Одобренные:</b>\n" + approvedText
	}
	if pendingText != "" {
		text += "\n<b>В ожидании:</b>\n" + pendingText
	}
	if rejectedText != "" {
		text += "\n<b>Отклонённые:</b>\n" + rejectedText
	}

	return text
}
