package view

import (
	"activity-bot/internal/helpers"
	"activity-bot/internal/model"
	"fmt"
	"time"
)

func FormatRestSet(member model.ChatMember, date time.Time, reason string) string {
	text := fmt.Sprintf("Участник %s %s в рест до %s",
		helpers.RoleLink(member),
		helpers.Gendered(member.User.Gender, "добавлен", "добавлена"),
		helpers.FormatToHumanDateTime(date),
	)
	if reason != "" {
		text += fmt.Sprintf("\n\nПричина: %s", reason)
	}
	return text
}

func FormatRestRequest(user model.ChatMember, date time.Time, reason string) string {
	text := fmt.Sprintf(
		"Для участника %s запрошен рест до %s",
		helpers.RoleLink(user),
		helpers.FormatToHumanDateTime(date),
	)
	if reason != "" {
		text += fmt.Sprintf("\n\nПричина: %s", reason)
	}
	return text
}

func FormatRestShow(m model.ChatMember) string {
	if m.RestUntil.IsZero() {
		return fmt.Sprintf("%s не находится в ресте", helpers.RoleLink(m))
	}
	message := "%s находится в ресте до %s"
	if m.RestUntil.Before(time.Now()) {
		message = "Рест %s был завершен %s"
	}
	text := fmt.Sprintf(message, helpers.RoleLink(m), helpers.FormatToHumanDateTime(m.RestUntil))
	if m.RestReason != "" {
		text += fmt.Sprintf("\n\nПричина: %s", m.RestReason)
	}
	return text
}

func FormatRestEnded(user model.ChatMember, isSelf bool) string {
	if isSelf {
		return "Вы успешно удалены из реста"
	}
	return fmt.Sprintf("Участник %s успешно %s из реста",
		helpers.RoleLink(user),
		helpers.Gendered(user.User.Gender, "удалён", "удалена"),
	)
}

func FormatRestNotInRest(user model.ChatMember, isSelf bool) string {
	if isSelf {
		return "Вы не находитесь в ресте"
	}
	return fmt.Sprintf("Пользователь %s не находится в ресте", helpers.RoleLink(user))
}

func FormatRestRequestApproved(user model.ChatMember, restUntil time.Time) string {
	return fmt.Sprintf("Запрос одобрен. У %s рест до %s", helpers.RoleLink(user), helpers.FormatToHumanDateTime(restUntil))
}

func FormatRestRequestRejected(user *model.ChatMember) string {
	if user == nil {
		return "Запрос на рест отклонён"
	}
	return fmt.Sprintf("Запрос на рест для %s отклонён", helpers.RoleLink(*user))
}

func FormatRestRequests(requests []model.ApprovedRestRequest) string {
	if len(requests) == 0 {
		return "Список рестов пуст"
	}

	var approvedText, rejectedText string
	var cm model.ChatMember
	for i, r := range requests {
		cm = r.ChatMember
		reasonPart := ""
		if r.Reason != "" {
			reasonPart = fmt.Sprintf(" (%s)", r.Reason)
		}
		timePart := fmt.Sprintf("• Запрошено %s", helpers.FormatToHumanDateTime(r.RequestedAt))
		if !r.UpdatedAt.IsZero() {
			timePart += fmt.Sprintf("\n• Одобрено %s", helpers.FormatToHumanDateTime(r.RequestedAt))
		}
		line := fmt.Sprintf("<code>%d</code> %s Срок окончания %s%s\n%s\n\n", i+1, isRestActiveMessage(r), helpers.FormatToHumanDateTime(r.RestUntil), reasonPart, timePart)

		switch r.Status {
		case "approved":
			approvedText += line
		case "rejected":
			rejectedText += line
		}
	}

	text := fmt.Sprintf("Список рестов %s:\n", helpers.RoleLink(cm))

	if approvedText != "" {
		text += "\nОдобренные:<blockquote expandable>" + approvedText + "</blockquote>"
	}
	if rejectedText != "" {
		text += "\nОтклонённые:<blockquote expandable>" + rejectedText + "</blockquote>"
	}

	return text + "\nЧтобы удалить определенный рест введите команду <code>удалить рест @участник номер</code>"
}

func isRestActiveMessage(rr model.ApprovedRestRequest) string {
	if !rr.ChatMember.IsRestActive(time.Now()) || !rr.ChatMember.RestUntil.Equal(rr.RestUntil) {
		return fmt.Sprintf("%s Недействителен\n", helpers.DangerEmoji())
	}
	return fmt.Sprintf("%s Действителен\n", helpers.SuccessEmoji())

}
